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

## Development Guidelines

When working on this project:
- Maintain the FastAPI backend structure
- Follow React/Vite conventions for frontend development
- Ensure API integrations remain functional
- Keep the feed content engaging and digestible
