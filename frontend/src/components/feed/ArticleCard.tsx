import React from 'react'
import { ExternalLink, FileText } from 'lucide-react'
import { Link } from '@tanstack/react-router'

interface ArticleCardProps {
  id?: number
  title: string
  summary: string
  source_url: string
  published_at: string
  document_number?: string | null
}

export const ArticleCard: React.FC<ArticleCardProps> = ({
  title,
  summary,
  source_url,
  published_at: _published_at,
  document_number,
}) => {

  // Generate a placeholder image based on title hash
  const colors = ['from-blue-500 to-cyan-500', 'from-purple-500 to-pink-500', 'from-orange-500 to-red-500', 'from-green-500 to-emerald-500']
  const colorIndex = title.charCodeAt(0) % colors.length
  const bgGradient = colors[colorIndex]

  return (
    <div className="rounded-lg overflow-hidden bg-white hover:shadow-lg transition-shadow">
      {/* Image Placeholder */}
      <div className={`w-full h-40 bg-gradient-to-br ${bgGradient} flex items-center justify-center text-white text-4xl font-bold opacity-90`}>
        📄
      </div>
      
      {/* Content */}
      <div className="p-4">
        <h3 className="text-lg font-bold mb-2 line-clamp-2 text-gray-900">
          {title}
        </h3>
        <p className="text-sm text-gray-600 mb-4 line-clamp-2">{summary}</p>
        <div className="flex gap-3">
          {document_number && (
            <Link
              to="/articles/$documentNumber"
              params={{ documentNumber: document_number }}
              className="inline-flex items-center gap-1 text-sm font-semibold text-blue-600 hover:text-blue-800"
            >
              <FileText className="w-4 h-4" />
              View Details
            </Link>
          )}
          <a
            href={source_url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-sm font-semibold text-gray-600 hover:text-gray-800"
            aria-label="Read More"
          >
            <ExternalLink className="w-4 h-4" />
            Federal Register
          </a>
        </div>
      </div>
    </div>
  )
}
