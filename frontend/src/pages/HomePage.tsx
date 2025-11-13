import { Link } from '@tanstack/react-router'

export default function HomePage() {
  return (
    <div className="space-y-12">
      <section className="text-center space-y-4 py-12">
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight">
          Stay Informed About Your Government
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Get daily updates about Federal Register announcements, policy changes, and government actions.
          Stay up-to-date with what matters to you.
        </p>
        <div className="pt-4">
          <Link
            to="/feed"
            className="inline-block px-8 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition"
          >
            View Latest Updates
          </Link>
        </div>
      </section>

      <section className="grid md:grid-cols-3 gap-8 py-12">
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">Always Updated</h3>
          <p className="text-muted-foreground">
            Get real-time Federal Register updates every 15 minutes.
          </p>
        </div>
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">AI Summaries</h3>
          <p className="text-muted-foreground">
            Complex government documents summarized by AI for easy understanding.
          </p>
        </div>
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">Share Updates</h3>
          <p className="text-muted-foreground">
            Share important announcements with your network easily.
          </p>
        </div>
      </section>
    </div>
  )
}
