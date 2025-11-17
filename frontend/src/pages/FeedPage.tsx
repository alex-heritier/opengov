import { useState } from 'react'
import { FeedList } from '../components/feed/FeedList'
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-react'

export default function FeedPage() {
  const [searchQuery, setSearchQuery] = useState('')

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

        {/* Section Header */}
        <div className="space-y-4 sm:space-y-6">
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900">Latest Updates</h2>
          <FeedList />
        </div>
      </div>
    </div>
  )
}
