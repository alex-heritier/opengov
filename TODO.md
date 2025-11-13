# OpenGov TODO List

**NOTE:** All tasks must be simple and actionable. If a task cannot be completed in a single context window (10-20 minutes, <= 200k conversation tokens), it must be broken down into subtasks.

**SCOPE:** This TODO covers Phase 1 only - Basic Federal Register scraping + Grok AI processing + website to create viral buzz.

## Phase 1: Development Environment & Tools

- [ ] Create Makefile for development workflow
  - [ ] Add `make install` (install backend & frontend deps)
  - [ ] Add `make run` (start backend & frontend dev servers)
  - [ ] Add `make test` (run backend + frontend tests)
  - [ ] Add `make build` (build frontend for production)

- [ ] Verify development environment
  - [ ] Check Python 3.11+ is installed
  - [ ] Check Node.js 18+ is installed
  - [ ] Document required versions in README

- [ ] Set up basic logging
  - [ ] Configure Python logging in backend
  - [ ] Add request/response logging middleware
  - [ ] Add error logging

## Phase 2: Backend Foundation

- [ ] Create backend app package
  - [ ] Create `backend/app/__init__.py`
  - [ ] Create `backend/app/main.py` with `FastAPI()` instance
  - [ ] Add `GET /health` endpoint returning `{"status": "ok"}`

- [ ] Create backend requirements
  - [ ] Add `fastapi`, `uvicorn[standard]`
  - [ ] Add `sqlalchemy`, `alembic`
  - [ ] Add `httpx`
  - [ ] Add `apscheduler`
  - [ ] Add `python-dotenv`
  - [ ] Add `pytest`, `pytest-asyncio` (dev)

- [ ] Configure database and SQLAlchemy
  - [ ] Create `backend/app/database.py` with engine, `SessionLocal`, `Base`
  - [ ] Create `backend/app/models/__init__.py`
  - [ ] Initialize Alembic: `cd backend && alembic init alembic`
  - [ ] Configure `backend/alembic.ini` with SQLite URL
  - [ ] Update `backend/alembic/env.py` to import `Base` from `app.database`

- [ ] Set up environment configuration
  - [ ] Create `backend/app/config.py` to load environment variables
  - [ ] Create `backend/.env.example` with required variables (GROK_API_KEY, DATABASE_URL, FEDERAL_REGISTER_API_URL)

## Phase 3: Backend Data Models

- [ ] Implement Article model
  - [ ] Create `backend/app/models/article.py`
  - [ ] Fields: id, title, summary, source_url, published_at, created_at, updated_at
  - [ ] Add indexes on published_at for sorting

- [ ] Implement FederalRegister model
  - [ ] Create `backend/app/models/federal_register.py`
  - [ ] Store raw Federal Register entries
  - [ ] Fields: id, document_number, raw_data (JSON), fetched_at, processed

- [ ] Create Pydantic schemas
  - [ ] Create `backend/app/schemas/__init__.py`
  - [ ] Create `backend/app/schemas/article.py` (ArticleResponse, ArticleDetail)
  - [ ] Create `backend/app/schemas/feed.py` (FeedResponse with pagination)

- [ ] Generate initial migration
  - [ ] Run `alembic revision --autogenerate -m "Initial models"`
  - [ ] Run `alembic upgrade head`

## Phase 4: External API Services

- [ ] Validate API access
  - [ ] Test Federal Register API endpoint with curl/httpx
  - [ ] Test Grok API with sample request
  - [ ] Document rate limits for both APIs

- [ ] Implement Federal Register service
  - [ ] Create `backend/app/services/__init__.py`
  - [ ] Create `backend/app/services/federal_register.py`
  - [ ] Implement `async fetch_recent_documents(days=1)` function
  - [ ] Add timeout handling (30s default)
  - [ ] Add retry logic with exponential backoff
  - [ ] Respect Federal Register API rate limits
  - [ ] Add error logging

- [ ] Implement Grok service
  - [ ] Create `backend/app/services/grok.py`
  - [ ] Implement `async summarize_text(text: str) -> str` with Grok API
  - [ ] Add prompt engineering for viral, engaging summaries
  - [ ] Add timeout handling (60s for AI processing)
  - [ ] Add error handling and fallback behavior
  - [ ] Add token usage logging

## Phase 5: Background Scraper Worker

- [ ] Implement scraper worker
  - [ ] Create `backend/app/workers/__init__.py`
  - [ ] Create `backend/app/workers/scraper.py`
  - [ ] Add `async fetch_and_process()` function:
    - Fetch new Federal Register items
    - Check for duplicates in database
    - Summarize with Grok API
    - Insert Article rows
  - [ ] Add error handling with logging
  - [ ] Add deduplication logic

- [ ] Configure scheduler
  - [ ] Add APScheduler configuration to run every 15 minutes
  - [ ] Add startup event in `main.py` to start scheduler
  - [ ] Add shutdown event to stop scheduler gracefully
  - [ ] Add logging for each scraper run
  - [ ] Add health check endpoint for scraper status

## Phase 6: Backend API Endpoints

- [ ] Add CORS middleware
  - [ ] Configure `CORSMiddleware` in `main.py`
  - [ ] Allow frontend origin (localhost:5173 for dev)
  - [ ] Allow credentials, methods="*", headers="*"

- [ ] Create routers package
  - [ ] Create `backend/app/routers/__init__.py`

- [ ] Implement feed router
  - [ ] Create `backend/app/routers/feed.py`
  - [ ] Add `GET /api/feed` (paginated articles)
    - Query params: page (default 1), limit (default 20), sort (newest/oldest)
    - Return articles with pagination metadata
  - [ ] Add `GET /api/feed/{article_id}` (single article details)
  - [ ] Add proper error responses (404, 500)
  - [ ] Include all routers in main.py

- [ ] Add admin endpoints (optional)
  - [ ] Create `backend/app/routers/admin.py`
  - [ ] Add `POST /api/admin/scrape` (manual trigger)
  - [ ] Add `GET /api/admin/stats` (article count, last scrape time)

## Phase 7: Frontend Foundation

- [ ] Initialize frontend app
  - [ ] Run `npm create vite@latest frontend -- --template react-ts`
  - [ ] Run `cd frontend && npm install`
  - [ ] Verify `src/main.tsx` and `src/App.tsx` exist

- [ ] Configure frontend dependencies
  - [ ] Install `@tanstack/react-router @tanstack/react-router-devtools`
  - [ ] Install `@tanstack/react-query @tanstack/react-query-devtools`
  - [ ] Install `zustand`
  - [ ] Install `axios`
  - [ ] Run `npx shadcn@latest init` (configure Tailwind)
  - [ ] Install shadcn components: `card`, `button`, `skeleton`, `badge`

- [ ] Set up environment configuration
  - [ ] Create `frontend/.env.example` with `VITE_API_URL=http://localhost:8000`
  - [ ] Create `frontend/.env.local` (gitignored) with actual values

## Phase 8: Frontend Routing & Layout

- [ ] Set up routing structure
  - [ ] Create `frontend/src/routes/__root.tsx`
  - [ ] Create `frontend/src/routes/index.tsx` (landing/home page)
  - [ ] Create `frontend/src/routes/feed.tsx` (main feed page)
  - [ ] Create `frontend/src/routes/article.$id.tsx` (single article view)
  - [ ] Configure router in `main.tsx`

- [ ] Create layout components
  - [ ] Create `frontend/src/components/layout/Header.tsx`
    - App logo/title
    - Navigation links
  - [ ] Create `frontend/src/components/layout/Footer.tsx`
    - Basic footer info
  - [ ] Use layout in `__root.tsx`

- [ ] Add metadata for viral sharing
  - [ ] Update `index.html` with Open Graph tags
  - [ ] Add Twitter Card tags
  - [ ] Add favicon and app icons
  - [ ] Add proper page titles and descriptions

## Phase 9: Frontend API Integration

- [ ] Set up API client
  - [ ] Create `frontend/src/api/client.ts` (axios instance)
  - [ ] Configure base URL from environment
  - [ ] Add request/response interceptors
  - [ ] Add error handling

- [ ] Create API query hooks
  - [ ] Create `frontend/src/api/queries.ts`
  - [ ] Add `useFeedQuery(page, limit)` with TanStack Query
  - [ ] Add `useArticleQuery(id)` with TanStack Query
  - [ ] Configure QueryClient in `main.tsx`

- [ ] Create state stores
  - [ ] Create `frontend/src/stores/feedStore.ts`
  - [ ] Add feed preferences (sort order, page size)

## Phase 10: Frontend UI Components

- [ ] Implement feed UI
  - [ ] Create `frontend/src/components/feed/ArticleCard.tsx`
    - Display title, summary, date
    - Link to full article
    - Link to source
    - Use shadcn Card component
  - [ ] Create `frontend/src/components/feed/FeedList.tsx`
    - Map over articles
    - Display ArticleCard for each
    - Show loading state
    - Show error state
    - Show empty state

- [ ] Add loading and error states
  - [ ] Create `frontend/src/components/ui/LoadingSpinner.tsx`
  - [ ] Create `frontend/src/components/ui/ErrorMessage.tsx`
  - [ ] Use shadcn Skeleton for ArticleCard loading state
  - [ ] Add ErrorBoundary for component errors

- [ ] Implement pagination
  - [ ] Add pagination controls to FeedList
  - [ ] Add "Load More" button or infinite scroll
  - [ ] Show current page and total pages

- [ ] Add mobile responsiveness
  - [ ] Test all components on mobile viewport
  - [ ] Ensure readable font sizes
  - [ ] Ensure touch targets are 44px minimum
  - [ ] Test on iOS Safari and Chrome mobile

## Phase 11: Frontend Pages

- [ ] Build landing page
  - [ ] Create hero section with app description
  - [ ] Add "View Feed" call-to-action button
  - [ ] Add basic styling

- [ ] Build feed page
  - [ ] Connect FeedList to useFeedQuery
  - [ ] Add sorting controls (newest/oldest)
  - [ ] Add page title "Latest Government Updates"
  - [ ] Add auto-refresh indicator (if implemented)

- [ ] Build article detail page
  - [ ] Connect to useArticleQuery
  - [ ] Display full article details
  - [ ] Add "Back to Feed" link
  - [ ] Add "View Source" link (Federal Register)
  - [ ] Add share button (native Web Share API)

## Phase 12: Testing

- [ ] Set up backend testing
  - [ ] Create `backend/pytest.ini` with asyncio mode
  - [ ] Create `backend/tests/__init__.py`
  - [ ] Add test for `GET /health`
  - [ ] Add tests for `GET /api/feed` with mocked database
  - [ ] Add tests for `GET /api/feed/{id}` (success and 404)
  - [ ] Add tests for Federal Register service (mocked httpx)
  - [ ] Add tests for Grok service (mocked API)

- [ ] Set up frontend testing
  - [ ] Install `vitest`, `@testing-library/react`, `@testing-library/user-event`, `jsdom`
  - [ ] Create `frontend/vitest.config.ts`
  - [ ] Create `frontend/src/test/setup.ts`
  - [ ] Write test for ArticleCard (mock data)
  - [ ] Write test for FeedList (mock query)
  - [ ] Run `npm test` to verify

## Phase 13: Documentation

- [ ] Create data model documentation
  - [ ] Create `docs/model.md`
  - [ ] Document Article schema with fields
  - [ ] Document FederalRegister schema with fields
  - [ ] Add ER diagram or ASCII representation
  - [ ] Keep in sync with SQLAlchemy models

- [ ] Create API documentation
  - [ ] Create `docs/api.md`
  - [ ] Document `GET /health`
  - [ ] Document `GET /api/feed` with parameters and response
  - [ ] Document `GET /api/feed/{id}` with response examples
  - [ ] Document error response format

- [ ] Create UI style guide
  - [ ] Create `docs/style.md`
  - [ ] Document color system
  - [ ] Document typography scale
  - [ ] Document component patterns
  - [ ] Document responsive breakpoints

- [ ] Update README
  - [ ] Add project overview and mission statement
  - [ ] Add setup instructions (prerequisites, installation, running)
  - [ ] Add tech stack section
  - [ ] Add testing instructions
  - [ ] Add screenshot placeholder
  - [ ] Add license

## Phase 14: Polish & Pre-Launch

- [ ] Security basics
  - [ ] Add input validation to all endpoints (max lengths)
  - [ ] Add rate limiting middleware (slowapi or similar)
  - [ ] Ensure CORS is configured correctly
  - [ ] Review for SQL injection protection (SQLAlchemy handles)
  - [ ] Add basic request size limits

- [ ] Performance basics
  - [ ] Add database indexes on frequently queried columns
  - [ ] Add HTTP caching headers to feed endpoint
  - [ ] Optimize Grok prompts for faster responses
  - [ ] Test with 100+ articles in database

- [ ] Monitoring & observability
  - [ ] Add structured logging (JSON logs)
  - [ ] Add endpoint for viewing recent scraper runs
  - [ ] Add basic error tracking
  - [ ] Add health check for database connection

- [ ] Browser testing
  - [ ] Test on Chrome, Firefox, Safari
  - [ ] Test on iOS Safari and Chrome Android
  - [ ] Verify Open Graph tags render correctly
  - [ ] Test Web Share API on mobile

- [ ] Deployment preparation
  - [ ] Document deployment requirements
  - [ ] Add production configuration examples
  - [ ] Test production build (`make build`)
  - [ ] Document environment variables for production
