import React from 'react'
import { ExternalLink, FileText, Bookmark, BookmarkCheck } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { Button } from '@/components/ui/button'
import { useToggleBookmarkMutation } from '@/api/queries'
import { useAuthStore } from '@/stores/authStore'

interface ArticleCardProps {
  id?: number
  title: string
  summary: string
  source_url: string
  published_at: string
  document_number?: string | null
  is_bookmarked?: boolean
}

export const ArticleCard: React.FC<ArticleCardProps> = ({
  id,
  title,
  summary,
  source_url,
  published_at: _published_at,
  document_number,
  is_bookmarked = false,
}) => {
  const { isAuthenticated } = useAuthStore()
  const toggleBookmark = useToggleBookmarkMutation()

  // Sanitize summary to prevent XSS attacks
  const sanitizedSummary = DOMPurify.sanitize(summary, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p'],
    ALLOWED_ATTR: []
  })

  const handleToggleBookmark = async () => {
    if (!id) return
    await toggleBookmark.mutateAsync(id)
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
          )}
        </div>
      </div>
    </article>
  )
}
