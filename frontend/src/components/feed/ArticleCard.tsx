import React from 'react'
import { ExternalLink, FileText } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import DOMPurify from 'dompurify'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

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
  
  // Sanitize summary to prevent XSS attacks
  const sanitizedSummary = DOMPurify.sanitize(summary, { 
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'br', 'p'],
    ALLOWED_ATTR: []
  })

  return (
    <Card className="h-full flex flex-col hover:shadow-lg transition-shadow">
      {/* Image Placeholder */}
      <div className={`w-full h-32 sm:h-40 bg-gradient-to-br ${bgGradient} flex items-center justify-center text-white text-2xl sm:text-4xl font-bold opacity-90`}>
        📄
      </div>
      
      {/* Content */}
      <CardContent className="p-3 sm:p-4 flex flex-col flex-1">
        <h3 className="text-base sm:text-lg font-bold mb-2 line-clamp-2 text-foreground">
          {title}
        </h3>
        <p className="text-xs sm:text-sm text-muted-foreground mb-4 line-clamp-2" dangerouslySetInnerHTML={{ __html: sanitizedSummary }}></p>
      </CardContent>

      {/* Footer with Actions */}
      <CardFooter className="flex gap-2 p-3 sm:p-4 pt-0">
        {document_number && (
          <Button
            asChild
            variant="ghost"
            size="sm"
            className="text-xs sm:text-sm"
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
          className="text-xs sm:text-sm"
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
      </CardFooter>
    </Card>
  )
}
