import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useFeedQuery } from '@/api/queries'
import { useFeedStore } from '@/stores/feedStore'
import { ArticleCard } from './ArticleCard'

export const FeedList: React.FC = () => {
  const [page, setPage] = useState(1)
  const [allArticles, setAllArticles] = useState<any[]>([])
  const [hasMore, setHasMore] = useState(true)
  const { sort, pageSize } = useFeedStore()
  const { data, isLoading, error } = useFeedQuery(page, pageSize, sort)
  const loadingRef = useRef(false)

  // Accumulate articles from each page
  useEffect(() => {
    if (data?.articles && data.articles.length > 0) {
      setAllArticles((prev) => {
        const newArticles = data.articles.filter(
          (article) => !prev.some((a) => a.id === article.id)
        )
        return [...prev, ...newArticles]
      })
      setHasMore(data.has_next)
      loadingRef.current = false
    }
  }, [data])

  // Handle scroll with window scroll event
  const handleScroll = useCallback(() => {
    if (loadingRef.current || !hasMore || isLoading) return
    
    const scrollHeight = document.documentElement.scrollHeight
    const scrollTop = document.documentElement.scrollTop
    const clientHeight = window.innerHeight
    
    // Trigger when near bottom (300px)
    if (scrollTop + clientHeight >= scrollHeight - 300) {
      loadingRef.current = true
      setPage((p) => p + 1)
    }
  }, [hasMore, isLoading])

  useEffect(() => {
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [handleScroll])

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-300 rounded-lg text-red-700">
        Failed to load articles. Please try again later.
      </div>
    )
  }

  if (allArticles.length === 0 && !isLoading) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 text-lg">No articles found.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-3 gap-6">
        {allArticles.map((article) => (
          <ArticleCard
            key={article.id}
            id={article.id}
            title={article.title}
            summary={article.summary}
            source_url={article.source_url}
            published_at={article.published_at}
          />
        ))}
      </div>

      {/* Loading indicator */}
      {isLoading && (
        <div className="grid grid-cols-3 gap-6">
          {[...Array(3)].map((_, i) => (
            <div
              key={i}
              className="rounded-lg overflow-hidden bg-gray-200 animate-pulse h-72"
            />
          ))}
        </div>
      )}



      {/* End of feed message */}
      {!hasMore && allArticles.length > 0 && (
        <div className="text-center py-8">
          <p className="text-gray-500">No more articles to load.</p>
        </div>
      )}
    </div>
  )
}
