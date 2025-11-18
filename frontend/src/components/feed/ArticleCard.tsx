import React from 'react'
import { ExternalLink, FileText, Bookmark, BookmarkCheck, ThumbsUp, ThumbsDown } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { Button } from '@/components/ui/button'
import { useToggleBookmarkMutation, useToggleLikeMutation } from '@/api/queries'
import { useAuthStore } from '@/stores/authStore'

interface ArticleCardProps {
  id?: number
  title: string
  summary: string
  source_url: string
  published_at: string
  document_number?: string | null
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
  document_number,
  is_bookmarked = false,
  user_like_status = null,
  likes_count = 0,
  dislikes_count = 0,
}) => {
  const { isAuthenticated } = useAuthStore()
  const toggleBookmark = useToggleBookmarkMutation()
  const toggleLike = useToggleLikeMutation()

  // Sanitize summary to prevent XSS attacks
  const sanitizedSummary = DOMPurify.sanitize(summary, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p'],
    ALLOWED_ATTR: []
  })

  const handleToggleBookmark = async () => {
    if (!id) return
    await toggleBookmark.mutateAsync(id)
  }

  const handleLike = async () => {
    if (!id) return
    await toggleLike.mutateAsync({ articleId: id, isPositive: true })
  }

  const handleDislike = async () => {
    if (!id) return
    await toggleLike.mutateAsync({ articleId: id, isPositive: false })
  }

  return (
    <article className="border-b border-gray-200 py-4 sm:py-6 hover:bg-gray-50 transition-colors">
      {/* Content */}
      <div className="space-y-3">
        <h3 className="text-base sm:text-lg font-bold text-gray-900 leading-snug">
          {title}
        </h3>
        <p
          className="text-sm text-gray-600 line-clamp-3 leading-relaxed"
          dangerouslySetInnerHTML={{ __html: sanitizedSummary }}
        />

        {/* Actions */}
        <div className="flex flex-wrap gap-2 pt-2">
          {document_number && (
            <Button
              asChild
              variant="ghost"
              size="sm"
              className="text-xs sm:text-sm h-8 px-3"
            >
              <Link
                to="/articles/$documentNumber"
                params={{ documentNumber: document_number }}
              >
                <FileText className="w-4 h-4" />
                View Details
              </Link>
            </Button>
          )}
          <Button
            asChild
            variant="outline"
            size="sm"
            className="text-xs sm:text-sm h-8 px-3"
          >
            <a
              href={source_url}
              target="_blank"
              rel="noopener noreferrer"
              aria-label="Read on Federal Register"
            >
              <ExternalLink className="w-4 h-4" />
              Federal Register
            </a>
          </Button>
          {isAuthenticated && (
            <>
              <Button
                variant={user_like_status === true ? "default" : "outline"}
                size="sm"
                onClick={handleLike}
                disabled={toggleLike.isPending}
                className="text-xs sm:text-sm h-8 px-3"
                aria-label="Like article"
              >
                <ThumbsUp className="w-4 h-4" />
                {likes_count > 0 && <span className="ml-1">{likes_count}</span>}
              </Button>
              <Button
                variant={user_like_status === false ? "default" : "outline"}
                size="sm"
                onClick={handleDislike}
                disabled={toggleLike.isPending}
                className="text-xs sm:text-sm h-8 px-3"
                aria-label="Dislike article"
              >
                <ThumbsDown className="w-4 h-4" />
                {dislikes_count > 0 && <span className="ml-1">{dislikes_count}</span>}
              </Button>
              <Button
                variant={is_bookmarked ? "default" : "outline"}
                size="sm"
                onClick={handleToggleBookmark}
                disabled={toggleBookmark.isPending}
                className="text-xs sm:text-sm h-8 px-3"
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
              </Button>
            </>
          )}
        </div>
      </div>
    </article>
  )
}
