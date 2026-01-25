import { useMutation, useQueryClient } from "@tanstack/react-query";
import client from "@/api/client";
import {
  useArticleUIStore,
  type ArticleUIState,
} from "@/store/article-ui-store";

interface BookmarkResponse {
  is_bookmarked: boolean;
}

interface RemoveBookmarkResponse {
  success: boolean;
}

const defaultUI = (): ArticleUIState => ({
  is_bookmarked: false,
  user_like_status: null,
  likes_count: 0,
  dislikes_count: 0,
});

export function useToggleBookmarkMutation() {
  const queryClient = useQueryClient();
  const setBookmark = useArticleUIStore((s) => s.setBookmark);
  const restore = useArticleUIStore((s) => s.restore);
  const byId = useArticleUIStore((s) => s.byId);

  return useMutation({
    mutationFn: async (feedEntryId: number) => {
      const { data } = await client.post<BookmarkResponse>(
        `/api/bookmarks/${feedEntryId}`,
        {},
      );
      return data;
    },

    onMutate: async (feedEntryId) => {
      const prev = byId[feedEntryId] ?? defaultUI();
      setBookmark(feedEntryId, !prev.is_bookmarked);
      return { feedEntryId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.feedEntryId, ctx.prev);
    },

    onSuccess: (_data, _vars) => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
}

export function useRemoveBookmarkMutation() {
  const queryClient = useQueryClient();
  const setBookmark = useArticleUIStore((s) => s.setBookmark);
  const restore = useArticleUIStore((s) => s.restore);
  const byId = useArticleUIStore((s) => s.byId);

  return useMutation({
    mutationFn: async (feedEntryId: number) => {
      const { data } = await client.delete<RemoveBookmarkResponse>(
        `/api/bookmarks/${feedEntryId}`,
      );
      return data;
    },

    onMutate: async (feedEntryId) => {
      const prev = byId[feedEntryId] ?? defaultUI();
      setBookmark(feedEntryId, false);
      return { feedEntryId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.feedEntryId, ctx.prev);
    },

    onSuccess: (_data, _feedEntryId) => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
}
