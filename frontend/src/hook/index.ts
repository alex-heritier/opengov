// Domain hooks - primary API for consumers
export { useFeed } from "./useFeed";
export { useBookmarks } from "./useBookmarks";
export { useFeedEntry } from "./useFeedEntry";
export { useProfile } from "./useProfile";
export type { UserUpdate } from "./useProfile";
export { useAuth } from "./useAuth";
export { useAuthRefresh } from "./useAuthRefresh";

// Environment hook
export { useEnvironment } from "./useEnvironment";
export type { Environment } from "./useEnvironment";

// Types
export type {
  User,
  FeedEntry,
  FeedEntryResponse,
  BookmarkedEntry,
} from "./types";

// Feed entry view hook (merges server data with UI store)
export { useFeedEntryView } from "./useFeedEntryView";

// Low-level query hooks (for advanced use cases)
export {
  useFeedQuery,
  useFeedEntryQuery,
  useBookmarksQuery,
  useToggleBookmarkMutation,
  useRemoveBookmarkMutation,
  useToggleLikeMutation,
  useRemoveLikeMutation,
  useUpdateProfileMutation,
} from "@/query";
export type { FeedResponse, FeedEntryDetail } from "@/query";

// Low-level store hooks (for advanced use cases)
export { useFeedStore } from "@/store/feedStore";
export { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";
