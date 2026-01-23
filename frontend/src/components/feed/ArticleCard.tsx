import React from 'react'
import { ExternalLink, FileText, Bookmark, BookmarkCheck, ThumbsUp, ThumbsDown } from 'lucide-react'
import { Link, useNavigate } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { useToggleBookmarkMutation, useToggleLikeMutation } from '@/hooks'
import { useAuthStore } from '@/stores/authStore'

interface ArticleCardProps {
  id?: number
  title: string
  summary: string
  source_url: string
  published_at: string
  unique_key?: string | null
  is_bookmarked?: boolean
  user_like_status?: boolean | null
  likes_count?: number
  dislikes_count?: number
}

export const ArticleCard: React.FC<ArticleCardProps> = ({
  id,
  title,
  summary,
  source_url,
  published_at: _published_at,
  unique_key,
  is_bookmarked = false,
  user_like_status = null,
  likes_count = 0,
  dislikes_count = 0,
}) => {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuthStore()
  const toggleBookmark = useToggleBookmarkMutation()
  const toggleLike = useToggleLikeMutation()

  // Single optimistic state object
  const [optimistic, setOptimistic] = React.useState({
    bookmarked: is_bookmarked,
    likeStatus: user_like_status,
    likesCount: likes_count,
    dislikesCount: dislikes_count,
  })

  // Sync with props when query refetches
  React.useEffect(() => {
    setOptimistic({
      bookmarked: is_bookmarked,
      likeStatus: user_like_status,
      likesCount: likes_count,
      dislikesCount: dislikes_count,
    })
  }, [is_bookmarked, user_like_status, likes_count, dislikes_count])

  const requireAuth = () => {
    if (!isAuthenticated) {
      navigate({ to: '/login' })
      return false
    }
    return true
  }

  const sanitizedSummary = DOMPurify.sanitize(summary, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p'],
    ALLOWED_ATTR: []
  })

  const handleToggleBookmark = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!requireAuth() || !id) return

    const prevState = optimistic.bookmarked
    setOptimistic((s) => ({ ...s, bookmarked: !s.bookmarked }))

    try {
      await toggleBookmark.mutateAsync(id)
    } catch (error) {
      console.error('Failed to toggle bookmark:', error)
      setOptimistic((s) => ({ ...s, bookmarked: prevState }))
    }
  }

  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!requireAuth() || !id) return

    const prevState = { ...optimistic }
    const isCurrentlyLiked = optimistic.likeStatus === true
    const isCurrentlyDisliked = optimistic.likeStatus === false

    // Toggle like (if already liked, clear; otherwise set to true)
    const newStatus = isCurrentlyLiked ? null : true
    setOptimistic((s) => {
      let likesCount = s.likesCount
      let dislikesCount = s.dislikesCount

      if (isCurrentlyLiked) {
        likesCount = Math.max(0, likesCount - 1)
      } else if (isCurrentlyDisliked) {
        likesCount += 1
        dislikesCount = Math.max(0, dislikesCount - 1)
      } else {
        likesCount += 1
      }

      return {
        ...s,
        likeStatus: newStatus,
        likesCount,
        dislikesCount,
      }
    })

    try {
      await toggleLike.mutateAsync({ articleId: id, isPositive: true })
    } catch (error) {
      console.error('Failed to like article:', error)
      setOptimistic(prevState)
    }
  }

  const handleDislike = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!requireAuth() || !id) return

    const prevState = { ...optimistic }
    const isCurrentlyLiked = optimistic.likeStatus === true
    const isCurrentlyDisliked = optimistic.likeStatus === false

    // Toggle dislike (if already disliked, clear; otherwise set to false)
    const newStatus = isCurrentlyDisliked ? null : false
    setOptimistic((s) => {
      let likesCount = s.likesCount
      let dislikesCount = s.dislikesCount

      if (isCurrentlyDisliked) {
        dislikesCount = Math.max(0, dislikesCount - 1)
      } else if (isCurrentlyLiked) {
        dislikesCount += 1
        likesCount = Math.max(0, likesCount - 1)
      } else {
        dislikesCount += 1
      }

      return {
        ...s,
        likeStatus: newStatus,
        likesCount,
        dislikesCount,
      }
    })

    try {
      await toggleLike.mutateAsync({ articleId: id, isPositive: false })
    } catch (error) {
      console.error('Failed to dislike article:', error)
      setOptimistic(prevState)
    }
  }

  return (
    <article className="border-b border-gray-200 py-4 sm:py-6 hover:bg-gray-50 transition-colors">
      <div className="space-y-3">
        <h3 className="text-base sm:text-lg font-bold text-gray-900 leading-snug">
          {unique_key ? (
            <Link
              to="/articles/$slug"
              params={{ slug: unique_key }}
              className="hover:underline hover:text-blue-700 transition-colors"
            >
              {title}
            </Link>
          ) : (
            title
          )}
        </h3>
        <p
          className="text-sm text-gray-600 line-clamp-3 leading-relaxed"
          dangerouslySetInnerHTML={{ __html: sanitizedSummary }}
        />

        <div className="flex flex-wrap gap-2 pt-2">
          {unique_key && (
            <Link
              to="/articles/$slug"
              params={{ slug: unique_key }}
              className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium bg-gray-100 hover:bg-gray-200 transition-colors text-gray-900 no-underline"
            >
              <FileText className="w-4 h-4" />
              View Details
            </Link>
          )}
          <a
            href={source_url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Read on Federal Register"
            className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium border border-gray-300 bg-white hover:bg-gray-50 transition-colors cursor-pointer text-gray-900 no-underline"
          >
            <ExternalLink className="w-4 h-4" />
            Federal Register
          </a>
          {isAuthenticated && (
            <>
              <button
                onClick={handleLike}
                disabled={toggleLike.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  optimistic.likeStatus === true
                    ? 'bg-green-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label="Like article"
              >
                <ThumbsUp className="w-4 h-4" />
                {optimistic.likesCount > 0 && <span>{optimistic.likesCount}</span>}
              </button>
              <button
                onClick={handleDislike}
                disabled={toggleLike.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  optimistic.likeStatus === false
                    ? 'bg-red-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label="Dislike article"
              >
                <ThumbsDown className="w-4 h-4" />
                {optimistic.dislikesCount > 0 && <span>{optimistic.dislikesCount}</span>}
              </button>
              <button
                onClick={handleToggleBookmark}
                disabled={toggleBookmark.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  optimistic.bookmarked
                    ? 'bg-blue-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label={optimistic.bookmarked ? "Remove bookmark" : "Bookmark article"}
              >
                {optimistic.bookmarked ? (
                  <>
                    <BookmarkCheck className="w-4 h-4" />
                    Bookmarked
                  </>
                ) : (
                  <>
                    <Bookmark className="w-4 h-4" />
                    Bookmark
                  </>
                )}
              </button>
            </>
          )}
        </div>
      </div>
    </article>
  )
}
