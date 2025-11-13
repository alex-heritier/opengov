# OpenGov TODO List

**NOTE:** All tasks must be simple and actionable. If a task cannot be completed in a single context window (10-20 minutes, <= 200k conversation tokens), it must be broken down into subtasks.

**SCOPE:** This TODO covers Phase 1 only - Basic Federal Register scraping + Grok AI processing + website to create viral buzz.

## Phase 1: Development Environment & Tools

- [x] Create Makefile for development workflow
  - [x] Add `make install` (install backend & frontend deps)
  - [x] Add `make run` (start backend & frontend dev servers)
  - [x] Add `make test` (run backend + frontend tests)
  - [x] Add `make build` (build frontend for production)

- [x] Verify development environment
  - [x] Check Python 3.11+ is installed
  - [x] Check Node.js 18+ is installed
  - [x] Document required versions in README

- [x] Set up basic logging
  - [x] Configure Python logging in backend
  - [x] Add request/response logging middleware
  - [x] Add error logging

## Phase 2: Backend Foundation

- [x] Create backend app package
  - [x] Create `backend/app/__init__.py`
  - [x] Create `backend/app/main.py` with `FastAPI()` instance
  - [x] Add `GET /health` endpoint returning `{"status": "ok"}`

- [x] Create backend requirements
  - [x] Add `fastapi`, `uvicorn[standard]`
  - [x] Add `sqlalchemy`, `alembic`
  - [x] Add `httpx`
  - [x] Add `apscheduler`
  - [x] Add `python-dotenv`
  - [x] Add `pytest`, `pytest-asyncio` (dev)

- [x] Configure database and SQLAlchemy
  - [x] Create `backend/app/database.py` with engine, `SessionLocal`, `Base`
  - [x] Create `backend/app/models/__init__.py`
  - [x] Initialize Alembic: `cd backend && alembic init alembic`
  - [x] Configure `backend/alembic.ini` with SQLite URL
  - [x] Update `backend/alembic/env.py` to import `Base` from `app.database`

- [x] Set up environment configuration
  - [x] Create `backend/app/config.py` to load environment variables
  - [x] Create `backend/.env.example` with required variables (GROK_API_KEY, DATABASE_URL, FEDERAL_REGISTER_API_URL)

## Phase 3: Backend Data Models

- [x] Implement Article model
  - [x] Create `backend/app/models/article.py`
  - [x] Fields: id, title, summary, source_url, published_at, created_at, updated_at
  - [x] Add indexes on published_at for sorting

- [x] Implement FederalRegister model
  - [x] Create `backend/app/models/federal_register.py`
  - [x] Store raw Federal Register entries
  - [x] Fields: id, document_number, raw_data (JSON), fetched_at, processed

- [x] Create Pydantic schemas
  - [x] Create `backend/app/schemas/__init__.py`
  - [x] Create `backend/app/schemas/article.py` (ArticleResponse, ArticleDetail)
  - [x] Create `backend/app/schemas/feed.py` (FeedResponse with pagination)

- [x] Generate initial migration
  - [x] Run `alembic revision --autogenerate -m "Initial models"`
  - [x] Run `alembic upgrade head`

## Phase 4: External API Services

- [x] Validate API access
  - [x] Test Federal Register API endpoint with curl/httpx
  - [x] Test Grok API with sample request
  - [x] Document rate limits for both APIs

- [x] Implement Federal Register service
  - [x] Create `backend/app/services/__init__.py`
  - [x] Create `backend/app/services/federal_register.py`
  - [x] Implement `async fetch_recent_documents(days=1)` function
  - [x] Add timeout handling (30s default)
  - [x] Add retry logic with exponential backoff
  - [x] Respect Federal Register API rate limits
  - [x] Add error logging

- [x] Implement Grok service
  - [x] Create `backend/app/services/grok.py`
  - [x] Implement `async summarize_text(text: str) -> str` with Grok API
  - [x] Add prompt engineering for viral, engaging summaries
  - [x] Add timeout handling (60s for AI processing)
  - [x] Add error handling and fallback behavior
  - [x] Add token usage logging

## Phase 5: Background Scraper Worker

- [x] Implement scraper worker
  - [x] Create `backend/app/workers/__init__.py`
  - [x] Create `backend/app/workers/scraper.py`
  - [x] Add `async fetch_and_process()` function:
     - Fetch new Federal Register items
     - Check for duplicates in database
     - Summarize with Grok API
     - Insert Article rows
  - [x] Add error handling with logging
  - [x] Add deduplication logic

- [x] Configure scheduler
  - [x] Add APScheduler configuration to run every 15 minutes
  - [x] Add startup event in `main.py` to start scheduler
  - [x] Add shutdown event to stop scheduler gracefully
  - [x] Add logging for each scraper run
  - [x] Add health check endpoint for scraper status

## Phase 6: Backend API Endpoints

- [x] Add CORS middleware
  - [x] Configure `CORSMiddleware` in `main.py`
  - [x] Allow frontend origin (localhost:5173 for dev)
  - [x] Allow credentials, methods="*", headers="*"

- [x] Create routers package
  - [x] Create `backend/app/routers/__init__.py`

- [x] Implement feed router
  - [x] Create `backend/app/routers/feed.py`
  - [x] Add `GET /api/feed` (paginated articles)
    - Query params: page (default 1), limit (default 20), sort (newest/oldest)
    - Return articles with pagination metadata
  - [x] Add `GET /api/feed/{article_id}` (single article details)
  - [x] Add proper error responses (404, 500)
  - [x] Include all routers in main.py

- [x] Add admin endpoints (optional)
  - [x] Create `backend/app/routers/admin.py`
  - [x] Add `POST /api/admin/scrape` (manual trigger)
  - [x] Add `GET /api/admin/stats` (article count, last scrape time)

## Phase 7: Frontend Foundation

- [x] Initialize frontend app
  - [x] Run `npm create vite@latest frontend -- --template react-ts`
  - [x] Run `cd frontend && npm install`
  - [x] Verify `src/main.tsx` and `src/App.tsx` exist

- [x] Configure frontend dependencies
  - [x] Install `@tanstack/react-router @tanstack/react-router-devtools`
  - [x] Install `@tanstack/react-query @tanstack/react-query-devtools`
  - [x] Install `zustand`
  - [x] Install `axios`
  - [x] Run `npx shadcn@latest init` (configure Tailwind)
  - [x] Install shadcn components: `card`, `button`, `skeleton`, `badge`

- [x] Set up environment configuration
  - [x] Create `frontend/.env.example` with `VITE_API_URL=http://localhost:8000`
  - [x] Create `frontend/.env.local` (gitignored) with actual values

## Phase 8: Frontend Routing & Layout

- [x] Set up routing structure
  - [x] Create `frontend/src/routes/__root.tsx`
  - [x] Create `frontend/src/routes/index.tsx` (landing/home page)
  - [x] Create `frontend/src/routes/feed.tsx` (main feed page)
  - [x] Create `frontend/src/routes/article.$id.tsx` (single article view)
  - [x] Configure router in `main.tsx`

- [x] Create layout components
  - [x] Create `frontend/src/components/layout/Header.tsx`
    - App logo/title
    - Navigation links
  - [x] Create `frontend/src/components/layout/Footer.tsx`
    - Basic footer info
  - [x] Use layout in `__root.tsx`

- [x] Add metadata for viral sharing
  - [x] Update `index.html` with Open Graph tags
  - [x] Add Twitter Card tags
  - [x] Add favicon and app icons
  - [x] Add proper page titles and descriptions

## Phase 9: Frontend API Integration

- [x] Set up API client
  - [x] Create `frontend/src/api/client.ts` (axios instance)
  - [x] Configure base URL from environment
  - [x] Add request/response interceptors
  - [x] Add error handling

- [x] Create API query hooks
  - [x] Create `frontend/src/api/queries.ts`
  - [x] Add `useFeedQuery(page, limit)` with TanStack Query
  - [x] Add `useArticleQuery(id)` with TanStack Query
  - [x] Configure QueryClient in `main.tsx`

- [x] Create state stores
  - [x] Create `frontend/src/stores/feedStore.ts`
  - [x] Add feed preferences (sort order, page size)

## Phase 10: Frontend UI Components

- [x] Implement feed UI
  - [x] Create `frontend/src/components/feed/ArticleCard.tsx`
    - Display title, summary, date
    - Link to full article
    - Link to source
    - Use shadcn Card component
  - [x] Create `frontend/src/components/feed/FeedList.tsx`
    - Map over articles
    - Display ArticleCard for each
    - Show loading state
    - Show error state
    - Show empty state

- [x] Add loading and error states
  - [x] Create `frontend/src/components/ui/LoadingSpinner.tsx`
  - [x] Create `frontend/src/components/ui/ErrorMessage.tsx`
  - [x] Use shadcn Skeleton for ArticleCard loading state
  - [x] Add ErrorBoundary for component errors

- [x] Implement pagination
  - [x] Add pagination controls to FeedList
  - [x] Add "Load More" button or infinite scroll
  - [x] Show current page and total pages

- [x] Add mobile responsiveness
  - [x] Test all components on mobile viewport
  - [x] Ensure readable font sizes
  - [x] Ensure touch targets are 44px minimum
  - [x] Test on iOS Safari and Chrome mobile

## Phase 11: Frontend Pages

- [x] Build landing page
  - [x] Create hero section with app description
  - [x] Add "View Feed" call-to-action button
  - [x] Add basic styling

- [x] Build feed page
  - [x] Connect FeedList to useFeedQuery
  - [x] Add sorting controls (newest/oldest)
  - [x] Add page title "Latest Government Updates"
  - [x] Add auto-refresh indicator (if implemented)

- [x] Build article detail page
  - [x] Connect to useArticleQuery
  - [x] Display full article details
  - [x] Add "Back to Feed" link
  - [x] Add "View Source" link (Federal Register)
  - [x] Add share button (native Web Share API)

## Phase 12: Testing

- [x] Set up backend testing
  - [x] Create `backend/pytest.ini` with asyncio mode
  - [x] Create `backend/tests/__init__.py`
  - [x] Add test for `GET /health`
  - [x] Add tests for `GET /api/feed` with mocked database
  - [x] Add tests for `GET /api/feed/{id}` (success and 404)
  - [x] Add tests for Federal Register service (mocked httpx)
  - [x] Add tests for Grok service (mocked API)

- [x] Set up frontend testing
  - [x] Install `vitest`, `@testing-library/react`, `@testing-library/user-event`, `jsdom`
  - [x] Create `frontend/vitest.config.ts`
  - [x] Create `frontend/src/test/setup.ts`
  - [x] Write test for ArticleCard (mock data)
  - [x] Write test for FeedList (mock query)
  - [x] Run `npm test` to verify

## Phase 13: Documentation

- [x] Create data model documentation
  - [x] Create `docs/model.md`
  - [x] Document Article schema with fields
  - [x] Document FederalRegister schema with fields
  - [x] Add ER diagram or ASCII representation
  - [x] Keep in sync with SQLAlchemy models

- [x] Create API documentation
  - [x] Create `docs/api.md`
  - [x] Document `GET /health`
  - [x] Document `GET /api/feed` with parameters and response
  - [x] Document `GET /api/feed/{id}` with response examples
  - [x] Document error response format

- [x] Create UI style guide
  - [x] Create `docs/style.md`
  - [x] Document color system
  - [x] Document typography scale
  - [x] Document component patterns
  - [x] Document responsive breakpoints

- [x] Update README
  - [x] Add project overview and mission statement
  - [x] Add setup instructions (prerequisites, installation, running)
  - [x] Add tech stack section
  - [x] Add testing instructions
  - [x] Add screenshot placeholder
  - [x] Add license

## Phase 14: Polish & Pre-Launch

- [x] Security basics
  - [x] Add input validation to all endpoints (max lengths)
  - [x] Add rate limiting middleware (slowapi)
    - 100 req/min for feed endpoints
    - 50 req/min for stats/scraper-runs
    - 10 req/min for manual scrape
  - [x] Ensure CORS is configured correctly
  - [x] Review for SQL injection protection (SQLAlchemy handles)
  - [x] Add basic request size limits (10 MB)

- [x] Performance basics
  - [x] Add database indexes on frequently queried columns (already done in Phases 3-5)
  - [x] Add HTTP caching headers to feed endpoint (Cache-Control, ETag)
  - [x] Optimize Grok prompts for faster responses (already done)
  - [x] Test with 100+ articles in database

- [x] Monitoring & observability
  - [x] Add request/response logging middleware with duration
  - [x] Add endpoint for viewing recent scraper runs (`GET /api/admin/scraper-runs`)
  - [x] Add scraper run tracking in database (ScraperRun model)
  - [x] Add health check for database connection (`GET /health/db`)

- [x] Browser testing
  - [x] Test on Chrome, Firefox, Safari
  - [x] Test on iOS Safari and Chrome Android
  - [x] Verify Open Graph tags render correctly
  - [x] Test Web Share API on mobile

- [x] Deployment preparation
  - [x] Document deployment requirements
  - [x] Add production configuration examples
  - [x] Test production build (`make build`)
  - [x] Document environment variables for production
