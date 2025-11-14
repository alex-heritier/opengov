import logging
import httpx
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from app.config import settings

logger = logging.getLogger(__name__)

# Constants for Grok API calls
GROK_MODEL = "grok-4-fast"
GROK_TEMPERATURE = 0.7
GROK_MAX_TOKENS = 300
SUMMARY_MAX_FALLBACK = 200

# Prompt for generating viral, engaging summaries
VIRAL_SUMMARY_PROMPT = """You are an expert at writing engaging, viral-worthy \
summaries of government documents and Federal Register entries.

Your task is to create a short, punchy summary (1-2 sentences max) that \
captures the essence of what the government is doing and why it matters to \
everyday Americans.

Guidelines:
- Be clear and accessible (avoid jargon)
- Focus on human impact
- Make it engaging and interesting
- Keep it under 280 characters when possible
- Start with the most important information

Document to summarize:
{text}

Generate only the summary, nothing else."""


@retry(
    stop=stop_after_attempt(2),  # Fewer retries for Grok (costs money)
    wait=wait_exponential(multiplier=1, min=1, max=5),
    retry=retry_if_exception_type((httpx.TimeoutException, httpx.HTTPStatusError)),
    reraise=True
)
async def _summarize_text_real(text: str) -> str:
    """
    Summarize text using Grok API.

    Args:
        text: Text to summarize

    Returns:
        Summary text, or original text if API fails
    """
    # Early return if text is empty
    if not text or not text.strip():
        logger.debug("Empty text provided for summarization, returning default")
        return "No summary available."

    # Early return if API key is not configured
    if not settings.GROK_API_KEY.strip():
        logger.warning("GROK_API_KEY not configured, returning truncated text")
        return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text

    try:
        prompt = VIRAL_SUMMARY_PROMPT.format(text=text)

        async with httpx.AsyncClient(timeout=settings.GROK_TIMEOUT) as client:
            response = await client.post(
                f"{settings.GROK_API_URL}/chat/completions",
                headers={
                    "Authorization": f"Bearer {settings.GROK_API_KEY}",
                    "Content-Type": "application/json",
                },
                json={
                    "model": GROK_MODEL,
                    "messages": [{"role": "user", "content": prompt}],
                    "temperature": GROK_TEMPERATURE,
                    "max_tokens": GROK_MAX_TOKENS,
                },
            )

            response.raise_for_status()
            data = response.json()

            summary = data.get("choices", [{}])[0].get("message", {}).get("content", "")

            if summary:
                logger.info(f"Successfully generated summary ({len(summary)} chars)")
                return summary.strip()
            else:
                logger.warning("Empty response from Grok API")
                return (
                    text[:SUMMARY_MAX_FALLBACK] + "..."
                    if len(text) > SUMMARY_MAX_FALLBACK else text
                )

    except httpx.TimeoutException:
        logger.warning(f"Grok API timeout after {settings.GROK_TIMEOUT}s, using truncated text")
        return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text
    except httpx.HTTPError as e:
        logger.warning(f"Grok API HTTP error: {e}, using truncated text")
        return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text
    except Exception as e:
        logger.warning(f"Error calling Grok API: {e}, using truncated text")
        return text[:SUMMARY_MAX_FALLBACK] + "..." if len(text) > SUMMARY_MAX_FALLBACK else text


def _get_summarizer():
    """Factory function to select the appropriate summarizer based on config."""
    if settings.USE_MOCK_GROK:
        logger.info("Using mock Grok summarizer for development")
        from app.services.grok_mock import summarize_text as mock_summarize
        return mock_summarize
    else:
        logger.info("Using real Grok API summarizer")
        return _summarize_text_real


# Initialize the summarizer at module load time
summarize_text = _get_summarizer()
