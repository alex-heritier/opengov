import { useQuery } from '@tanstack/react-query'
import client from '../api/client'
import { Article } from './types'

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
