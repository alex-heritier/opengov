import { useQuery } from "@tanstack/react-query";
import client from "@/api/client";
import { FeedEntryResponse } from "@/hook/types";
import { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";
import { useStoreWithEqualityFn } from "zustand/traditional";

export interface FeedEntryDetail extends FeedEntryResponse {
  updated_at: string;
}

export function useFeedEntryQuery(id: number) {
  const hydrate = useStoreWithEqualityFn(useFeedEntryUIStore, (s) => s.hydrate);

  return useQuery({
    queryKey: ["feedEntry", id],
    queryFn: async () => {
      const { data } = await client.get<FeedEntryResponse>(`/api/feed/${id}`);
      hydrate([data]);
      return data;
    },
    enabled: !!id,
    staleTime: 1000 * 60 * 10,
  });
}
