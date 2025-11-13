import { useQuery } from '@tanstack/react-query'
import client from './client'

interface Article {
  id: number
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
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
