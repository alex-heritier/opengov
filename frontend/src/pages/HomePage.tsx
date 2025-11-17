import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

export default function HomePage() {
  return (
    <div className="w-full max-w-6xl mx-auto">
      <section className="text-center space-y-4 sm:space-y-6 py-8 sm:py-12 px-4">
        <h1 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight text-gray-900">
          Stay Informed About Your Government
        </h1>
        <p className="text-base sm:text-lg md:text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
          Get daily updates about Federal Register announcements, policy changes, and government actions.
          Stay up-to-date with what matters to you.
        </p>
        <div className="pt-4 sm:pt-6">
          <Button asChild size="lg" className="text-sm sm:text-base h-11 sm:h-12 px-6 sm:px-8">
            <Link to="/feed">
              View Latest Updates
            </Link>
          </Button>
        </div>
      </section>

      <section className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 sm:gap-8 py-8 sm:py-12 px-4">
        <div className="space-y-3 text-center sm:text-left">
          <h3 className="text-lg sm:text-xl font-semibold text-gray-900">Always Updated</h3>
          <p className="text-sm sm:text-base text-gray-600 leading-relaxed">
            Get real-time Federal Register updates every 15 minutes.
          </p>
        </div>
        <div className="space-y-3 text-center sm:text-left">
          <h3 className="text-lg sm:text-xl font-semibold text-gray-900">AI Summaries</h3>
          <p className="text-sm sm:text-base text-gray-600 leading-relaxed">
            Complex government documents summarized by AI for easy understanding.
          </p>
        </div>
        <div className="space-y-3 text-center sm:text-left sm:col-span-2 md:col-span-1">
          <h3 className="text-lg sm:text-xl font-semibold text-gray-900">Share Updates</h3>
          <p className="text-sm sm:text-base text-gray-600 leading-relaxed">
            Share important announcements with your network easily.
          </p>
        </div>
      </section>
    </div>
  )
}
