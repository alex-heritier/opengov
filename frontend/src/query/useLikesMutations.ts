import { useMutation, useQueryClient } from "@tanstack/react-query";
import client from "@/api/client";
import { useArticleUIStore, type LikeStatus } from "@/store/article-ui-store";

interface LikeResponse {
  value: number;
}

interface RemoveLikeResponse {
  success: boolean;
}

export function useToggleLikeMutation() {
  const queryClient = useQueryClient();
  const applyReaction = useArticleUIStore((s) => s.applyReaction);
  const restore = useArticleUIStore((s) => s.restore);

  return useMutation({
    mutationFn: async ({
      feedEntryId,
      isPositive,
    }: {
      feedEntryId: number;
      isPositive: boolean;
    }) => {
      const { data } = await client.post<LikeResponse>(
        `/api/likes/${feedEntryId}`,
        {
          value: isPositive ? 1 : -1,
        },
      );
      return data;
    },

    onMutate: async ({ feedEntryId, isPositive }) => {
      const next: LikeStatus = isPositive ? true : false;
      const { prev } = applyReaction(feedEntryId, next);
      return { feedEntryId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.feedEntryId, ctx.prev);
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
}

export function useRemoveLikeMutation() {
  const queryClient = useQueryClient();
  const applyReaction = useArticleUIStore((s) => s.applyReaction);
  const restore = useArticleUIStore((s) => s.restore);

  return useMutation({
    mutationFn: async (feedEntryId: number) => {
      const { data } = await client.delete<RemoveLikeResponse>(
        `/api/likes/${feedEntryId}`,
      );
      return data;
    },

    onMutate: async (feedEntryId) => {
      const { prev } = applyReaction(feedEntryId, null);
      return { feedEntryId, prev };
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) restore(ctx.feedEntryId, ctx.prev);
    },

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] });
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
}
