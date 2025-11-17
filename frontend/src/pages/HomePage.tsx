import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

export default function HomePage() {
  return (
    <div className="space-y-8 sm:space-y-12">
      <section className="text-center space-y-3 sm:space-y-4 py-8 sm:py-12">
        <h1 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight">
          Stay Informed About Your Government
        </h1>
        <p className="text-base sm:text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto px-4">
          Get daily updates about Federal Register announcements, policy changes, and government actions.
          Stay up-to-date with what matters to you.
        </p>
        <div className="pt-3 sm:pt-4">
          <Button asChild size="lg" className="text-sm sm:text-base">
            <Link to="/feed">
              View Latest Updates
            </Link>
          </Button>
        </div>
      </section>

      <section className="grid grid-cols-1 md:grid-cols-3 gap-6 sm:gap-8 py-8 sm:py-12">
        <div className="space-y-2 px-4 sm:px-0">
          <h3 className="text-base sm:text-lg font-semibold">Always Updated</h3>
          <p className="text-sm sm:text-base text-muted-foreground">
            Get real-time Federal Register updates every 15 minutes.
          </p>
        </div>
        <div className="space-y-2 px-4 sm:px-0">
          <h3 className="text-base sm:text-lg font-semibold">AI Summaries</h3>
          <p className="text-sm sm:text-base text-muted-foreground">
            Complex government documents summarized by AI for easy understanding.
          </p>
        </div>
        <div className="space-y-2 px-4 sm:px-0">
          <h3 className="text-base sm:text-lg font-semibold">Share Updates</h3>
          <p className="text-sm sm:text-base text-muted-foreground">
            Share important announcements with your network easily.
          </p>
        </div>
      </section>
    </div>
  )
}
