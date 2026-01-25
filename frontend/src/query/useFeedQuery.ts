import { useInfiniteQuery } from "@tanstack/react-query";
import client from "@/api/client";
import { FeedResponse } from "@/hook/types";
import { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";
import { useStoreWithEqualityFn } from "zustand/traditional";

export function useFeedQuery(limit: number = 20, sort: string = "newest") {
  const hydrate = useStoreWithEqualityFn(useFeedEntryUIStore, (s) => s.hydrate);

  return useInfiniteQuery({
    queryKey: ["feed", limit, sort],
    queryFn: async ({ pageParam }) => {
      const { data } = await client.get<FeedResponse>("/api/feed", {
        params: { page: pageParam, limit, sort },
      });
      hydrate(data.items);
      return data;
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.has_next ? lastPage.page + 1 : undefined,
    staleTime: 1000 * 30,
    refetchOnWindowFocus: true,
  });
}
