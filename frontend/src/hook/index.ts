// Domain hooks - primary API for consumers
export { useFeed } from "./useFeed";
export { useBookmarks } from "./useBookmarks";
export { useArticle } from "./useArticle";
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
  Article,
  FeedEntryResponse,
  BookmarkedArticle,
} from "./types";

// Article view hook (merges server data with UI store)
export { useArticleView } from "./useArticleView";

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
export type { FeedResponse, ArticleDetail } from "@/query";

// Low-level store hooks (for advanced use cases)
export { useFeedStore } from "@/store/feedStore";
