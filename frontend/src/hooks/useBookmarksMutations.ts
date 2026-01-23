import { useMutation, useQueryClient } from '@tanstack/react-query'
import client from '../api/client'

export function useToggleBookmarkMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.post(`/api/bookmarks/${articleId}`, {})
      return data
    },
    onSuccess: () => {
      // Invalidate feed queries to update bookmark status
      queryClient.invalidateQueries({ queryKey: ['feed'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['bookmarks'] })
    },
  })
}

export function useRemoveBookmarkMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.delete(`/api/bookmarks/${articleId}`)
      return data
    },
    onSuccess: () => {
      // Invalidate queries to update UI
      queryClient.invalidateQueries({ queryKey: ['feed'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['bookmarks'] })
    },
  })
}
