/**
 * Domain hook for feed operations.
 * Orchestrates between feed query and feed store.
 */
import { useMemo } from 'react'
import { useFeedQuery } from '@/query'
import { useFeedStore } from '@/store/feedStore'

export function useFeed() {
  const { sort, pageSize, setSortOrder, setPageSize } = useFeedStore()
  const query = useFeedQuery(pageSize, sort)

  const articles = useMemo(
    () => query.data?.pages.flatMap((page) => page.articles) ?? [],
    [query.data]
  )

  const total = query.data?.pages[0]?.total ?? 0

  return {
    articles,
    total,
    hasNextPage: query.hasNextPage,
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
    fetchNextPage: query.fetchNextPage,

    sort,
    pageSize,
    setSortOrder,
    setPageSize,
  }
}
