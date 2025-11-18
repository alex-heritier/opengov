import { RouterProvider, RootRoute, Router, Route } from '@tanstack/react-router'
import RootLayout from './components/layout/RootLayout'
import FeedPage from './pages/FeedPage'
import HomePage from './pages/HomePage'
import AdminPage from './pages/AdminPage'
import ArticleDetailPage from './pages/ArticleDetailPage'
import AuthLoginPage from './pages/AuthLoginPage'
import AuthCallbackPage from './pages/AuthCallbackPage'
import BookmarksPage from './pages/BookmarksPage'
import ProfilePage from './pages/ProfilePage'
import LoginPage from './pages/LoginPage'

// Root route
const rootRoute = new RootRoute({
  component: RootLayout,
})

// Index route (home)
const indexRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

// Feed route
const feedRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/feed',
  component: FeedPage,
})

// Admin route
const adminRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/admin',
  component: AdminPage,
})

// Article detail route
const articleDetailRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/articles/$documentNumber',
  component: ArticleDetailPage,
})

// Auth login route
const authLoginRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/auth/login',
  component: AuthLoginPage,
})

// Auth callback route
const authCallbackRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/auth/callback',
  component: AuthCallbackPage,
})

// Bookmarks route
const bookmarksRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/bookmarks',
  component: BookmarksPage,
})

// Profile route
const profileRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/profile',
  component: ProfilePage,
})

// Login route
const loginRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})

// Create route tree
const routeTree = rootRoute.addChildren([
  indexRoute,
  feedRoute,
  adminRoute,
  articleDetailRoute,
  authLoginRoute,
  authCallbackRoute,
  bookmarksRoute,
  profileRoute,
  loginRoute,
])

// Create router
const router = new Router({ routeTree })

// Register router for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

export function App() {
  return <RouterProvider router={router} />
}
