import { useQuery } from '@tanstack/react-query'
import client from '../api/client'
import { BookmarkedArticle } from './types'

export function useBookmarksQuery() {
  return useQuery({
    queryKey: ['bookmarks'],
    queryFn: async () => {
      const { data } = await client.get<{ articles: BookmarkedArticle[] }>('/api/bookmarks')
      return data.articles
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  })
}
