/**
 * Domain hook for bookmark operations.
 * Orchestrates between bookmarks query and mutations.
 */
import { useBookmarksQuery, useToggleBookmarkMutation, useRemoveBookmarkMutation } from '@/query'

export function useBookmarks() {
  const query = useBookmarksQuery()
  const toggleMutation = useToggleBookmarkMutation()
  const removeMutation = useRemoveBookmarkMutation()

  return {
    // Query state
    bookmarks: query.data ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,

    // Mutations
    toggleBookmark: toggleMutation.mutate,
    toggleBookmarkAsync: toggleMutation.mutateAsync,
    isToggling: toggleMutation.isPending,

    removeBookmark: removeMutation.mutate,
    removeBookmarkAsync: removeMutation.mutateAsync,
    isRemoving: removeMutation.isPending,
  }
}
