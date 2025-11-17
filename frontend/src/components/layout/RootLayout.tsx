import { Outlet } from '@tanstack/react-router'
import Header from './Header'
import Footer from './Footer'

export default function RootLayout() {
  return (
    <div className="flex flex-col min-h-screen bg-background">
      <Header />
      <main className="flex-1 container mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
