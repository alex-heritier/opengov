import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { FeedList } from '../components/feed/FeedList'
import { Input } from '../components/ui/input'
import { Button } from '../components/ui/button'
import { Search, Bookmark, ThumbsUp } from 'lucide-react'
import { useAuthStore } from '../store/authStore'

export default function FeedPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const { isAuthenticated } = useAuthStore()

  return (
    <div className="w-full max-w-5xl mx-auto px-4 sm:px-6">
      {/* Main Content */}
      <div className="space-y-4 sm:space-y-6">
        {/* Search Bar */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            type="text"
            placeholder="Search Federal Register..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 h-10 sm:h-12 text-sm sm:text-base border-gray-300"
          />
        </div>

        {!isAuthenticated && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 sm:p-6">
            <h3 className="text-lg font-semibold text-blue-900 mb-2">Sign in to unlock all features</h3>
            <p className="text-sm text-blue-700 mb-4">
              Create a free account to bookmark articles, track what matters to you, and get personalized updates.
            </p>
            <div className="flex flex-wrap gap-3">
              <Button asChild size="sm">
                <Link to="/login">Sign In</Link>
              </Button>
              <Button asChild variant="outline" size="sm">
                <Link to="/feed">Browse as Guest</Link>
              </Button>
            </div>
            <div className="mt-4 pt-4 border-t border-blue-200 grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm text-blue-700">
              <div className="flex items-center gap-2">
                <Bookmark className="w-4 h-4" />
                <span>Save articles for later</span>
              </div>
              <div className="flex items-center gap-2">
                <ThumbsUp className="w-4 h-4" />
                <span>Track important issues</span>
              </div>
            </div>
          </div>
        )}

        {/* Section Header */}
        <div className="space-y-4 sm:space-y-6">
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900">Latest Updates</h2>
          <FeedList />
        </div>
      </div>
    </div>
  )
}
