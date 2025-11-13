# AI Agent Guidelines

## Overview

This is a government as a service (GaaS) project for corrupt countries with bad governance.

## Roadmap

1. Federal register scraper to create a viral buzz around us
2. Create the opengov GaaS product

## Tech Stack

- **Backend**: FastAPI
- **Frontend**: React with Vite
- **APIs**:
  - Federal Register API (periodically polled)
  - Grok's grok-4-fast API (for summarization and analysis)

## How It Works

The backend periodically hits the Federal Register API and runs the results through Grok's grok-4-fast API to summarize and analyze the register details. It then creates interesting, digestible little blurbs and articles on the `/feed` page for people to read and share.

## Implementation Notes

### Backend

- Use FastAPI's async/await for all API endpoints and external calls
- Implement periodic tasks with background workers (e.g., APScheduler or Celery)
- Store Federal Register data and Grok summaries in a database (use SQLAlchemy ORM)
- Create separate service modules for Federal Register API and Grok API integrations
- Include proper error handling and retry logic for external API calls
- Use environment variables for API keys and configuration
- Add rate limiting to prevent API quota exhaustion

### Frontend

- Structure components by feature (e.g., `/components/feed`, `/components/article`)
- Use React hooks for state management (Context API or Zustand for global state)
- Implement infinite scroll or pagination for the `/feed` page
- Make articles shareable with meta tags for social media previews
- Use Vite's environment variables for API endpoint configuration
- Keep UI responsive and mobile-friendly
- Add loading states and error boundaries for better UX

## Development Guidelines

When working on this project:
- Maintain the FastAPI backend structure
- Follow React/Vite conventions for frontend development
- Ensure API integrations remain functional
- Keep the feed content engaging and digestible
