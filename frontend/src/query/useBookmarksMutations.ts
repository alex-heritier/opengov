import { useMutation, useQueryClient } from "@tanstack/react-query";
import client from "@/api/client";
import {
  useArticleUIStore,
  type ArticleUIState,
} from "@/store/article-ui-store";

interface BookmarkResponse {
  article?: {
    id: number;
    is_bookmarked?: boolean;
    likes_count?: number;
    dislikes_count?: number;
    user_like_status?: boolean | null;
  };
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
  const hydrate = useArticleUIStore((s) => s.hydrate);
  const byId = useArticleUIStore((s) => s.byId);

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.post<BookmarkResponse>(
        `/api/bookmarks/${articleId}`,
        {},
      );
      return data;
    },

    onMutate: async (articleId) => {
      const prev = byId[articleId] ?? defaultUI();
      setBookmark(articleId, !prev.is_bookmarked);
      return { articleId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.articleId, ctx.prev);
    },

    onSuccess: (data, vars) => {
      if (data?.article) hydrate([data.article]);
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
      queryClient.invalidateQueries({ queryKey: ["article", vars] });
      queryClient.invalidateQueries({ queryKey: ["article", "slug"] });
    },
  });
}

export function useRemoveBookmarkMutation() {
  const queryClient = useQueryClient();
  const setBookmark = useArticleUIStore((s) => s.setBookmark);
  const restore = useArticleUIStore((s) => s.restore);
  const hydrate = useArticleUIStore((s) => s.hydrate);
  const byId = useArticleUIStore((s) => s.byId);

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.delete<BookmarkResponse>(
        `/api/bookmarks/${articleId}`,
      );
      return data;
    },

    onMutate: async (articleId) => {
      const prev = byId[articleId] ?? defaultUI();
      setBookmark(articleId, false);
      return { articleId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.articleId, ctx.prev);
    },

    onSuccess: (data, articleId) => {
      if (data?.article) hydrate([data.article]);
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
      queryClient.invalidateQueries({ queryKey: ["article", articleId] });
      queryClient.invalidateQueries({ queryKey: ["article", "slug"] });
    },
  });
}
