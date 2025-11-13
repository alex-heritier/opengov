import { Link } from '@tanstack/react-router'

export default function Header() {
  return (
    <header className="border-b border-border bg-card">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 hover:opacity-80">
          <span className="text-2xl font-bold text-primary">OpenGov</span>
        </Link>
        <nav className="flex gap-6">
          <Link to="/" className="text-sm font-medium hover:text-primary">
            Home
          </Link>
          <Link to="/feed" className="text-sm font-medium hover:text-primary">
            Feed
          </Link>
        </nav>
      </div>
    </header>
  )
}
