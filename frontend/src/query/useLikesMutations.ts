import { useMutation, useQueryClient } from "@tanstack/react-query";
import client from "@/api/client";
import { useArticleUIStore, type LikeStatus } from "@/store/article-ui-store";

interface LikeResponse {
  article?: {
    id: number;
    likes_count?: number;
    dislikes_count?: number;
    user_like_status?: boolean | null;
    is_bookmarked?: boolean;
  };
}

export function useToggleLikeMutation() {
  const queryClient = useQueryClient();
  const applyReaction = useArticleUIStore((s) => s.applyReaction);
  const restore = useArticleUIStore((s) => s.restore);
  const hydrate = useArticleUIStore((s) => s.hydrate);

  return useMutation({
    mutationFn: async ({
      articleId,
      isPositive,
    }: {
      articleId: number;
      isPositive: boolean;
    }) => {
      const { data } = await client.post<LikeResponse>(
        `/api/likes/${articleId}`,
        {
          is_positive: isPositive,
        },
      );
      return data;
    },

    onMutate: async ({ articleId, isPositive }) => {
      const next: LikeStatus = isPositive ? true : false;
      const { prev } = applyReaction(articleId, next);
      return { articleId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.articleId, ctx.prev);
    },

    onSuccess: (data, vars) => {
      if (data?.article) hydrate([data.article]);
      queryClient.invalidateQueries({ queryKey: ["article", vars.articleId] });
      queryClient.invalidateQueries({ queryKey: ["article", "slug"] });
    },
  });
}

export function useRemoveLikeMutation() {
  const queryClient = useQueryClient();
  const applyReaction = useArticleUIStore((s) => s.applyReaction);
  const restore = useArticleUIStore((s) => s.restore);
  const hydrate = useArticleUIStore((s) => s.hydrate);

  return useMutation({
    mutationFn: async (articleId: number) => {
      const { data } = await client.delete<LikeResponse>(
        `/api/likes/${articleId}`,
      );
      return data;
    },

    onMutate: async (articleId) => {
      const { prev } = applyReaction(articleId, null);
      return { articleId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.articleId, ctx.prev);
    },

    onSuccess: (data, articleId) => {
      if (data?.article) hydrate([data.article]);
      queryClient.invalidateQueries({ queryKey: ["article", articleId] });
      queryClient.invalidateQueries({ queryKey: ["article", "slug"] });
    },
  });
}
