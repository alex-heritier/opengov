import { Button } from '@/components/ui/button'
import { FlaskConical } from 'lucide-react'

export function TestLogin() {
  const handleTestLogin = () => {
    // Redirect to backend test login flow
    // Backend handles everything: user creation, cookie setting, redirect
    // This mimics the Google OAuth flow to avoid special cases
    window.location.href = `${import.meta.env.VITE_API_URL}/api/auth/test/login`
  }

  // Only show in development mode
  if (import.meta.env.PROD) {
    return null
  }

  return (
    <Button
      onClick={handleTestLogin}
      variant="outline"
      className="w-full flex items-center gap-3 border-dashed border-orange-300 text-orange-700 hover:bg-orange-50"
    >
      <FlaskConical className="w-5 h-5" />
      Test Login (Dev Only)
    </Button>
  )
}
