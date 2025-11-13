import { Outlet } from '@tanstack/react-router'
import Header from './Header'
import Footer from './Footer'

export default function RootLayout() {
  return (
    <div className="flex flex-col min-h-screen bg-background">
      <Header />
      <main className="flex-1 container mx-auto px-4 py-8">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
