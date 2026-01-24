# OpenGov TODO List

**NOTE:** All tasks must be simple and actionable. If a task cannot be completed in a single context window (10-20 minutes, <= 200k conversation tokens), it must be broken down into subtasks.

**SCOPE:** This TODO covers Phase 1 only - Basic Federal Register scraping + Grok AI processing + website to create viral buzz.

---

## TODOs

- [x] Implement Google OAuth signup/login
- [x] Create user profile page with logout button
- [x] Add like functionality for articles
- [x] Add bookmark/save functionality for articles
- [x] Implement article sharing (Facebook, X.com, copy link)
- [x] Create clickable articles on /feed page
- [x] Build dedicated /article page with query params (?q=xyz)
- [x] Display external source link on article page
- [x] Add AI summary to article page
- [x] Add AI keypoints to article page
- [x] Add partisan score meter (Democrat/Republican indicator)
- [x] Add AI impact score (low/medium/high)
- [x] Display agency on article page
- [ ] Fix OAuth state store concurrency issue (map not thread-safe)
- [ ] Fix N+1 query performance issue in assembler (bookmark/like bulk fetching)
  - Add BookmarkRepository.GetBookmarksForArticles(userID, articleIDs[])
  - Add LikeRepository.GetUserStatusesForArticles(userID, articleIDs[])
  - Add LikeRepository.GetCountsForArticles(articleIDs[])
  - Refactor assembler.go to use bulk methods instead of per-article queries
