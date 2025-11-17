import logging
import random

logger = logging.getLogger(__name__)

# Lorem ipsum snippets for mock summaries
LOREM_IPSUM_SUMMARIES = [
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. The government announced new regulations affecting federal operations and citizen engagement.",
    "Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. A new agency directive was issued to improve administrative processes.",
    "Ut enim ad minim veniam, quis nostrud exercitation ullamco. The Department released updated guidelines for public benefit programs.",
    "Duis aute irure dolor in reprehenderit in voluptate velit. Federal funding was allocated for infrastructure development initiatives.",
    "Excepteur sint occaecat cupidatat non proident, sunt in culpa. New environmental protection standards were established by the agency.",
    "Qui officia deserunt mollit anim id est laborum. The Treasury Department announced changes to tax reporting requirements.",
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod. A new rule was proposed affecting healthcare providers nationwide.",
    "Tempor incididunt ut labore et dolore magna aliqua ut enim. The agency expanded its public comment period for pending regulations.",
    "Ad minim veniam, quis nostrud exercitation ullamco laboris nisi. Federal grants were made available for research and development.",
    "Ut aliquip ex ea commodo consequat duis aute irure dolor. The government issued guidance on compliance with recent legislative changes.",
]


async def summarize_text(text: str) -> str:
    """
    Mock summarizer that returns Lorem Ipsum text.

    Use this for development to avoid API calls and costs.
    In production, use the real Grok summarizer.

    Args:
        text: Text to summarize (ignored in mock)

    Returns:
        Random Lorem Ipsum summary
    """
    if not text or not text.strip():
        logger.debug("Empty text provided for mock summarization")
        return "No summary available."

    summary = random.choice(LOREM_IPSUM_SUMMARIES)
    logger.info(f"Mock summarizer returning Lorem Ipsum ({len(summary)} chars)")
    return summary
