import { useMemo } from "react";
import type { FeedEntryResponse } from "./types";
import { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";
import { useStoreWithEqualityFn } from "zustand/traditional";

export function useFeedEntryView(entry: FeedEntryResponse): FeedEntryResponse {
  const ui = useStoreWithEqualityFn(useFeedEntryUIStore, (s) => s.byId[entry.id]);

  return useMemo(() => {
    if (!ui) return entry;
    return {
      ...entry,
      is_bookmarked: ui.is_bookmarked,
      user_like_status:
        ui.user_like_status === true
          ? 1
          : ui.user_like_status === false
            ? -1
            : null,
      likes_count: ui.likes_count,
      dislikes_count: ui.dislikes_count,
    };
  }, [entry, ui]);
}
