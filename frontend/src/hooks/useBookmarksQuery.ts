import { useQuery } from '@tanstack/react-query'
import client from '../api/client'
import { BookmarkedArticle } from './types'

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
