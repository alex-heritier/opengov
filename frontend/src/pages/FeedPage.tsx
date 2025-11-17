import { useState } from 'react'
import { FeedList } from '../components/feed/FeedList'
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-react'

export default function FeedPage() {
  const [searchQuery, setSearchQuery] = useState('')

  const sidebarItems = [
    { icon: '🏠', label: 'Home' },
    { icon: '📑', label: 'Home' },
    { icon: '🔍', label: 'Browse' },
    { icon: '🔎', label: 'Advanced Search' },
    { icon: 'ℹ️', label: 'About' },
    { icon: '✉️', label: 'Contact' },
    { icon: '❓', label: 'Help' },
  ]

  return (
    <div className="flex flex-col lg:flex-row gap-4 sm:gap-6 lg:gap-8">
      {/* Sidebar - Hidden on mobile, shown on lg+ screens */}
      <aside className="hidden lg:block w-32 flex-shrink-0">
        <nav className="space-y-6">
          {sidebarItems.map((item, idx) => (
            <div key={idx} className="flex items-center gap-3 text-sm cursor-pointer hover:text-blue-600">
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </div>
          ))}
        </nav>
      </aside>

      {/* Main Content */}
      <div className="flex-1 space-y-4 sm:space-y-6">
        {/* Search Bar */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search Federal Register..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 h-10 sm:h-12 text-sm sm:text-base"
          />
        </div>

        {/* Section Header */}
        <div className="space-y-4 sm:space-y-6">
          <h2 className="text-xl sm:text-2xl font-bold">Updates</h2>
          <FeedList />
        </div>
      </div>
    </div>
  )
}
