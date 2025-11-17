import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import client from './client'

export interface Article {
  id: number
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
  is_bookmarked?: boolean
  document_number?: string
  user_like_status?: boolean | null  // null = no vote, true = liked, false = disliked
  likes_count?: number
  dislikes_count?: number
}

export interface BookmarkedArticle {
  id: number
  document_number: string
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
  bookmarked_at: string
}

interface FeedResponse {
  articles: Article[]
  page: number
  limit: number
  total: number
  has_next: boolean
}

export function useFeedQuery(page: number = 1, limit: number = 20, sort: string = 'newest') {
  return useQuery({
    queryKey: ['feed', page, limit, sort],
    queryFn: async () => {
      const { data } = await client.get<FeedResponse>('/api/feed', {
        params: { page, limit, sort },
      })
      return data
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

export function useArticleQuery(id: number) {
  return useQuery({
    queryKey: ['article', id],
    queryFn: async () => {
      const { data } = await client.get<Article>(`/api/feed/${id}`)
      return data
    },
    enabled: !!id,
    staleTime: 1000 * 60 * 10, // 10 minutes
  })
}

// Bookmark queries and mutations
export function useBookmarksQuery() {
  return useQuery({
    queryKey: ['bookmarks'],
    queryFn: async () => {
      const { data } = await client.get<BookmarkedArticle[]>('/api/bookmarks')
      return data
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  })
}

export function useToggleBookmarkMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.post('/api/bookmarks/toggle', {
        frarticle_id: articleId,
      })
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

// Like queries and mutations
export function useToggleLikeMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ articleId, isPositive }: { articleId: number; isPositive: boolean }) => {
      const { data } = await client.post('/api/likes/toggle', {
        frarticle_id: articleId,
        is_positive: isPositive,
      })
      return data
    },
    onSuccess: () => {
      // Invalidate feed queries to update like status
      queryClient.invalidateQueries({ queryKey: ['feed'] })
      queryClient.invalidateQueries({ queryKey: ['article'] })
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
    },
  })
}
