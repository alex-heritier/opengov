import { Link } from '@tanstack/react-router'
import { useAuthStore } from '@/store/authStore'
import { useAuth } from '@/hook'
import { Bookmark, UserCircle2, LogOut } from 'lucide-react'

export default function Header() {
  const { isAuthenticated } = useAuthStore()
  const { logout } = useAuth()

  return (
    <header className="border-b border-gray-200 bg-white sticky top-0 z-50 shadow-sm">
      <div className="container mx-auto px-4 sm:px-6 py-3 sm:py-4 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity min-h-[44px]">
          <span className="text-xl sm:text-2xl font-bold text-gray-900">OpenGov</span>
        </Link>
        <nav className="flex gap-2 sm:gap-4">
          <Link
            to="/feed"
            className="text-sm sm:text-base font-medium text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors min-h-[44px] flex items-center"
          >
            Feed
          </Link>
          {isAuthenticated ? (
            <>
              <Link
                to="/bookmarks"
                className="text-sm sm:text-base font-medium text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors min-h-[44px] flex items-center gap-1"
              >
                <Bookmark className="w-4 h-4" />
                Bookmarks
              </Link>
              <Link
                to="/profile"
                className="text-sm sm:text-base font-medium text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors min-h-[44px] flex items-center gap-1"
              >
                <UserCircle2 className="w-4 h-4" />
                Account
              </Link>
              <button
                onClick={logout}
                className="text-sm sm:text-base font-medium text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors min-h-[44px] flex items-center gap-1"
              >
                <LogOut className="w-4 h-4" />
                Logout
              </button>
            </>
          ) : (
            <Link
              to="/login"
              className="text-sm sm:text-base font-medium text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors min-h-[44px] flex items-center"
            >
              Sign In
            </Link>
          )}
        </nav>
      </div>
    </header>
  )
}
