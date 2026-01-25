import { useMemo } from "react";
import type { FeedEntryResponse } from "./types";
import { useArticleUIStore } from "@/store/article-ui-store";

export function useArticleView(article: FeedEntryResponse): FeedEntryResponse {
  const ui = useArticleUIStore((s) => s.byId[article.id]);

  return useMemo(() => {
    if (!ui) return article;
    return {
      ...article,
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
  }, [article, ui]);
}
