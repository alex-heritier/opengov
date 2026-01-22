import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'

export default function AuthLoginPage() {
  const navigate = useNavigate()

  useEffect(() => {
    navigate({ to: '/login', replace: true })
  }, [navigate])

  return null
}
