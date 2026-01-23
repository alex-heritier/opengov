// Domain hooks - primary API for consumers
export { useFeed } from "./useFeed";
export { useBookmarks } from "./useBookmarks";
export { useArticle } from "./useArticle";
export { useProfile } from "./useProfile";
export type { UserUpdate } from "./useProfile";

// Auth hooks
export { useAuth } from "./useAuth";
export { AuthProvider } from "./AuthProvider";
export type { AuthContextValue } from "./AuthProvider";

// Types
export type { User, Article, BookmarkedArticle } from "./types";

// Article view hook (merges server data with UI store)
export { useArticleView } from "./useArticleView";

// Low-level query hooks (for advanced use cases)
export {
  useFeedQuery,
  useArticleQuery,
  useArticleBySlugQuery,
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
