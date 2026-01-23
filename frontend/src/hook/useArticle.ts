/**
 * Domain hook for article operations.
 * Orchestrates between article query, likes, and bookmarks.
 */
import {
  useArticleQuery,
  useToggleLikeMutation,
  useRemoveLikeMutation,
  useToggleBookmarkMutation,
} from "@/query";

export function useArticle(id: number) {
  const query = useArticleQuery(id);
  const likeMutation = useToggleLikeMutation();
  const unlikeMutation = useRemoveLikeMutation();
  const bookmarkMutation = useToggleBookmarkMutation();

  const like = (isPositive: boolean) => {
    likeMutation.mutate({ articleId: id, isPositive });
  };

  const unlike = () => {
    unlikeMutation.mutate(id);
  };

  const toggleBookmark = () => {
    bookmarkMutation.mutate(id);
  };

  return {
    // Query state
    article: query.data ?? null,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,

    // Actions
    like,
    unlike,
    toggleBookmark,
    isLiking: likeMutation.isPending,
    isUnliking: unlikeMutation.isPending,
    isBookmarking: bookmarkMutation.isPending,
  };
}
