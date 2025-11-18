import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { GoogleLogin } from '../components/auth/GoogleLogin'
import { useAuth } from '../contexts/AuthContext'

export default function LoginPage() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: '/feed' })
    }
  }, [isAuthenticated, navigate])

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <div className="mb-8 text-center">
          <h1 className="mb-2 text-3xl font-bold text-gray-900">OpenGov</h1>
          <p className="text-gray-600">
            Stay informed about what your government is doing
          </p>
        </div>

        <div className="mb-6">
          <GoogleLogin />
        </div>

        <div className="text-center">
          <p className="text-sm text-gray-500">
            By signing in, you agree to our Terms of Service and Privacy Policy
          </p>
        </div>
      </div>
    </div>
  )
}
