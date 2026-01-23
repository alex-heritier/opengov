import { useMutation, useQueryClient } from '@tanstack/react-query'
import client from '../api/client'

export function useToggleLikeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ articleId, isPositive }: { articleId: number; isPositive: boolean }) => {
      const { data } = await client.post(`/api/likes/${articleId}`, {
        is_positive: isPositive,
      })
      return data
    },
    onSuccess: () => {
      // Invalidate feed queries to update like status
      queryClient.invalidateQueries({ queryKey: ['feed'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
      queryClient.invalidateQueries({ queryKey: ['bookmarks'] })
    },
  })
}

export function useRemoveLikeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.delete(`/api/likes/${articleId}`)
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
