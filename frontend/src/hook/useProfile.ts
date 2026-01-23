/**
 * Domain hook for profile operations.
 * Orchestrates profile mutations.
 */
import { useUpdateProfileMutation, UserUpdate } from '@/query'

export function useProfile() {
  const mutation = useUpdateProfileMutation()

  return {
    updateProfile: mutation.mutate,
    updateProfileAsync: mutation.mutateAsync,
    isUpdating: mutation.isPending,
    isError: mutation.isError,
    error: mutation.error,
  }
}

export type { UserUpdate }
