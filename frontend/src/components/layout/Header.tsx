import { Link } from '@tanstack/react-router'

export default function Header() {
  return (
    <header className="border-b border-border bg-card sticky top-0 z-50">
      <div className="container mx-auto px-4 py-3 sm:py-4 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 hover:opacity-80">
          <span className="text-xl sm:text-2xl font-bold text-primary">OpenGov</span>
        </Link>
        <nav className="flex gap-4 sm:gap-6">
          <Link to="/" className="text-xs sm:text-sm font-medium hover:text-primary">
            Home
          </Link>
          <Link to="/feed" className="text-xs sm:text-sm font-medium hover:text-primary">
            Feed
          </Link>
        </nav>
      </div>
    </header>
  )
}
