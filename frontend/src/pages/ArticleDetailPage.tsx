import { useEffect, useState } from 'react'
import { useParams, Link } from '@tanstack/react-router'
import { ArrowLeft, ExternalLink, Calendar, Clock, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import ShareButtons from '@/components/share/ShareButtons'

interface ArticleDetail {
  id: number
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
  updated_at: string
  document_number: string | null
  unique_key: string
}

export default function ArticleDetailPage() {
  const { slug } = useParams({ from: '/articles/$slug' })
  const [article, setArticle] = useState<ArticleDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchArticle = async () => {
      try {
        setLoading(true)
        setError(null)

        const response = await fetch(`http://localhost:8000/api/feed/slug/${slug}`)

        if (!response.ok) {
          throw new Error(`Article not found (${response.status})`)
        }

        const data = await response.json()
        setArticle(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load article')
      } finally {
        setLoading(false)
      }
    }

    fetchArticle()
  }, [slug])

  if (loading) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8 space-y-4">
        <Skeleton className="h-8 w-3/4" />
        <Skeleton className="h-4 w-1/4" />
        <Skeleton className="h-32 w-full" />
      </div>
    )
  }

  if (error || !article) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {error || 'Article not found'}
          </AlertDescription>
        </Alert>
        <Button asChild variant="outline" className="mt-4">
          <Link to="/feed" className="inline-flex items-center gap-2">
            <ArrowLeft className="w-4 h-4" />
            Back to Feed
          </Link>
        </Button>
      </div>
    )
  }

  const formattedPublishedDate = new Date(article.published_at).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const formattedTime = new Date(article.published_at).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
      {/* Back Button */}
      <Button asChild variant="ghost" className="mb-4 sm:mb-6 text-sm sm:text-base">
        <Link to="/feed" className="inline-flex items-center gap-2">
          <ArrowLeft className="w-4 h-4" />
          Back to Feed
        </Link>
      </Button>

      {/* Article Header */}
      <article className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        {/* Title Banner */}
        <div className="bg-gray-50 border-b border-gray-200 px-4 sm:px-8 py-4 sm:py-6">
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-3">{article.title}</h1>
          <div className="flex flex-wrap gap-3 sm:gap-4 text-gray-600 text-xs sm:text-sm">
            <div className="flex items-center gap-2">
              <Calendar className="w-4 h-4" />
              <span>{formattedPublishedDate}</span>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4" />
              <span>{formattedTime}</span>
            </div>
            {article.document_number && (
              <div className="flex items-center gap-2">
                <span className="font-semibold">Doc #:</span>
                <span className="font-mono text-xs">{article.document_number}</span>
              </div>
            )}
          </div>
        </div>

        {/* Article Content */}
        <div className="px-4 sm:px-8 py-4 sm:py-6 space-y-4 sm:space-y-6">
          {/* Summary */}
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 mb-3">Summary</h2>
            <p className="text-sm sm:text-base text-gray-700 leading-relaxed whitespace-pre-wrap">{article.summary}</p>
          </div>

          {/* Share Buttons */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200">
            <ShareButtons
              title={article.title}
              url={typeof window !== 'undefined' ? window.location.href : ''}
              summary={article.summary}
            />
          </div>

          {/* Source Link */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200">
            <Button asChild className="text-sm sm:text-base">
              <a
                href={article.source_url}
                target="_blank"
                rel="noopener noreferrer"
              >
                <ExternalLink className="w-4 h-4 sm:w-5 sm:h-5" />
                View Full Document on Federal Register
              </a>
            </Button>
          </div>

          {/* Metadata */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200 text-xs sm:text-sm text-gray-500">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
              <div>
                <span className="font-semibold">Created:</span>{' '}
                {new Date(article.created_at).toLocaleDateString()}
              </div>
              <div>
                <span className="font-semibold">Last Updated:</span>{' '}
                {new Date(article.updated_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
  )
}
