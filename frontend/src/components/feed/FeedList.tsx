import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useFeedQuery } from '@/api/queries'
import { useFeedStore } from '@/stores/feedStore'
import { ArticleCard } from './ArticleCard'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle } from 'lucide-react'

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
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load articles. Please try again later.
        </AlertDescription>
      </Alert>
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
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
        {allArticles.map((article) => (
          <ArticleCard
            key={article.id}
            id={article.id}
            title={article.title}
            summary={article.summary}
            source_url={article.source_url}
            published_at={article.published_at}
            document_number={article.document_number}
          />
        ))}
      </div>

      {/* Loading indicator */}
      {isLoading && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="rounded-lg overflow-hidden h-72 flex flex-col">
              <Skeleton className="w-full h-32 sm:h-40" />
              <div className="p-3 sm:p-4 flex flex-col flex-1 space-y-2">
                <Skeleton className="h-6 w-3/4" />
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-2/3" />
                <div className="mt-auto pt-4 flex gap-2">
                  <Skeleton className="h-8 w-24" />
                  <Skeleton className="h-8 w-32" />
                </div>
              </div>
            </div>
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
