import { useState } from 'react'
import { Twitter, Facebook, Linkedin, Mail, Link2, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ShareButtonsProps {
  title: string
  url: string
  summary?: string
}

export default function ShareButtons({ title, url, summary }: ShareButtonsProps) {
  const [copied, setCopied] = useState(false)

  const encodedTitle = encodeURIComponent(title)
  const encodedUrl = encodeURIComponent(url)
  const encodedSummary = summary ? encodeURIComponent(summary) : ''

  const shareLinks = {
    twitter: `https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`,
    facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`,
    linkedin: `https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`,
    email: `mailto:?subject=${encodedTitle}&body=${encodedSummary}%0A%0A${encodedUrl}`,
  }

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy link:', err)
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-sm font-semibold text-gray-700">Share this article:</h3>
      <div className="flex flex-wrap gap-2">
        {/* Twitter */}
        <Button
          asChild
          variant="outline"
          size="sm"
          className="flex items-center gap-2 hover:bg-sky-50 hover:border-sky-400 hover:text-sky-600"
        >
          <a
            href={shareLinks.twitter}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Share on Twitter"
          >
            <Twitter className="w-4 h-4" />
            <span className="hidden sm:inline">Twitter</span>
          </a>
        </Button>

        {/* Facebook */}
        <Button
          asChild
          variant="outline"
          size="sm"
          className="flex items-center gap-2 hover:bg-blue-50 hover:border-blue-400 hover:text-blue-600"
        >
          <a
            href={shareLinks.facebook}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Share on Facebook"
          >
            <Facebook className="w-4 h-4" />
            <span className="hidden sm:inline">Facebook</span>
          </a>
        </Button>

        {/* LinkedIn */}
        <Button
          asChild
          variant="outline"
          size="sm"
          className="flex items-center gap-2 hover:bg-blue-50 hover:border-blue-600 hover:text-blue-700"
        >
          <a
            href={shareLinks.linkedin}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Share on LinkedIn"
          >
            <Linkedin className="w-4 h-4" />
            <span className="hidden sm:inline">LinkedIn</span>
          </a>
        </Button>

        {/* Email */}
        <Button
          asChild
          variant="outline"
          size="sm"
          className="flex items-center gap-2 hover:bg-gray-50 hover:border-gray-400 hover:text-gray-700"
        >
          <a
            href={shareLinks.email}
            aria-label="Share via Email"
          >
            <Mail className="w-4 h-4" />
            <span className="hidden sm:inline">Email</span>
          </a>
        </Button>

        {/* Copy Link */}
        <Button
          variant="outline"
          size="sm"
          onClick={handleCopyLink}
          className="flex items-center gap-2 hover:bg-green-50 hover:border-green-400 hover:text-green-600"
          aria-label="Copy link to clipboard"
        >
          {copied ? (
            <>
              <Check className="w-4 h-4" />
              <span className="hidden sm:inline">Copied!</span>
            </>
          ) : (
            <>
              <Link2 className="w-4 h-4" />
              <span className="hidden sm:inline">Copy Link</span>
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
