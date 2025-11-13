import { RouterProvider, RootRoute, Router, Route } from '@tanstack/react-router'
import RootLayout from './components/layout/RootLayout'
import FeedPage from './pages/FeedPage'
import HomePage from './pages/HomePage'
import AdminPage from './pages/AdminPage'

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

// Create route tree
const routeTree = rootRoute.addChildren([indexRoute, feedRoute, adminRoute])

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
