import { useQuery } from "@tanstack/react-query";
import client from "@/api/client";
import { FeedEntryResponse } from "@/hook/types";
import { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";
import { useStoreWithEqualityFn } from "zustand/traditional";

export function useBookmarksQuery() {
  const hydrate = useStoreWithEqualityFn(useFeedEntryUIStore, (s) => s.hydrate);

  return useQuery({
    queryKey: ["bookmarks"],
    queryFn: async () => {
      const { data } = await client.get<{ items: FeedEntryResponse[] }>(
        "/api/bookmarks",
      );
      hydrate(data.items);
      return data.items;
    },
    staleTime: 1000 * 60 * 2,
  });
}
