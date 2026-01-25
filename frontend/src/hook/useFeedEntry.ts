/**
 * Domain hook for feed entry operations.
 * Orchestrates between feed entry query, likes, and bookmarks.
 */
import {
  useFeedEntryQuery,
  useToggleLikeMutation,
  useRemoveLikeMutation,
  useToggleBookmarkMutation,
} from "@/query";

export function useFeedEntry(id: number) {
  const query = useFeedEntryQuery(id);
  const likeMutation = useToggleLikeMutation();
  const unlikeMutation = useRemoveLikeMutation();
  const bookmarkMutation = useToggleBookmarkMutation();

  const like = (isPositive: boolean) => {
    likeMutation.mutate({ feedEntryId: id, isPositive });
  };

  const unlike = () => {
    unlikeMutation.mutate(id);
  };

  const toggleBookmark = () => {
    bookmarkMutation.mutate(id);
  };

  return {
    // Query state
    entry: query.data ?? null,
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
