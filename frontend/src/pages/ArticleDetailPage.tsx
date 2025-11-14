import { useEffect, useState } from 'react'
import { useParams, Link } from '@tanstack/react-router'
import { ArrowLeft, ExternalLink, Calendar, Clock } from 'lucide-react'

interface ArticleDetail {
  id: number
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
  updated_at: string
  document_number: string | null
}

export default function ArticleDetailPage() {
  const { documentNumber } = useParams({ from: '/articles/$documentNumber' })
  const [article, setArticle] = useState<ArticleDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchArticle = async () => {
      try {
        setLoading(true)
        setError(null)

        const response = await fetch(`http://localhost:8000/api/feed/document/${documentNumber}`)

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
  }, [documentNumber])

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-3/4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/4"></div>
          <div className="h-32 bg-gray-200 rounded"></div>
        </div>
      </div>
    )
  }

  if (error || !article) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h2 className="text-xl font-bold text-red-800 mb-2">Error Loading Article</h2>
          <p className="text-red-600">{error || 'Article not found'}</p>
          <Link to="/feed" className="inline-flex items-center gap-2 mt-4 text-blue-600 hover:text-blue-800">
            <ArrowLeft className="w-4 h-4" />
            Back to Feed
          </Link>
        </div>
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
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Back Button */}
      <Link
        to="/feed"
        className="inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Feed
      </Link>

      {/* Article Header */}
      <article className="bg-white rounded-lg shadow-lg overflow-hidden">
        {/* Title Banner */}
        <div className="bg-gradient-to-r from-blue-600 to-cyan-600 px-8 py-6">
          <h1 className="text-3xl font-bold text-white mb-3">{article.title}</h1>
          <div className="flex flex-wrap gap-4 text-white/90 text-sm">
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
                <span className="font-mono">{article.document_number}</span>
              </div>
            )}
          </div>
        </div>

        {/* Article Content */}
        <div className="px-8 py-6 space-y-6">
          {/* Summary */}
          <div>
            <h2 className="text-xl font-bold text-gray-900 mb-3">Summary</h2>
            <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{article.summary}</p>
          </div>

          {/* Source Link */}
          <div className="pt-6 border-t border-gray-200">
            <a
              href={article.source_url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition-colors"
            >
              <ExternalLink className="w-5 h-5" />
              View Full Document on Federal Register
            </a>
          </div>

          {/* Metadata */}
          <div className="pt-6 border-t border-gray-200 text-sm text-gray-500">
            <div className="grid grid-cols-2 gap-4">
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
