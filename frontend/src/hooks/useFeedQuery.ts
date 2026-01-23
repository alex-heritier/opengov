import { useQuery } from '@tanstack/react-query'
import client from '../api/client'
import { Article } from './types'

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
