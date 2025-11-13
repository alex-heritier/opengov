import { useState } from 'react'
import { FeedList } from '../components/feed/FeedList'

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
    <div className="flex gap-8">
      {/* Sidebar */}
      <aside className="w-32 flex-shrink-0">
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
      <div className="flex-1 space-y-6">
        {/* Search Bar */}
        <div className="relative">
          <svg className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search Federal Register..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 h-12 text-base border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* Section Header */}
        <div className="space-y-6">
          <h2 className="text-2xl font-bold">Updates</h2>
          <FeedList />
        </div>
      </div>
    </div>
  )
}
