import React from 'react'
import { ExternalLink, FileText, Bookmark, BookmarkCheck, ThumbsUp, ThumbsDown } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { useToggleBookmarkMutation, useToggleLikeMutation } from '@/api/queries'
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
  const { isAuthenticated } = useAuthStore()
  const toggleBookmark = useToggleBookmarkMutation()
  const toggleLike = useToggleLikeMutation()

  const sanitizedSummary = DOMPurify.sanitize(summary, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p'],
    ALLOWED_ATTR: []
  })

  const handleToggleBookmark = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!id) return
    await toggleBookmark.mutateAsync(id)
  }

  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!id) return
    await toggleLike.mutateAsync({ articleId: id, isPositive: true })
  }

  const handleDislike = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!id) return
    await toggleLike.mutateAsync({ articleId: id, isPositive: false })
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
                  user_like_status === true
                    ? 'bg-green-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label="Like article"
              >
                <ThumbsUp className="w-4 h-4" />
                {likes_count > 0 && <span>{likes_count}</span>}
              </button>
              <button
                onClick={handleDislike}
                disabled={toggleLike.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  user_like_status === false
                    ? 'bg-red-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label="Dislike article"
              >
                <ThumbsDown className="w-4 h-4" />
                {dislikes_count > 0 && <span>{dislikes_count}</span>}
              </button>
              <button
                onClick={handleToggleBookmark}
                disabled={toggleBookmark.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  is_bookmarked
                    ? 'bg-blue-600 text-white'
                    : 'border border-gray-300 bg-white hover:bg-gray-50'
                }`}
                aria-label={is_bookmarked ? "Remove bookmark" : "Bookmark article"}
              >
                {is_bookmarked ? (
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
