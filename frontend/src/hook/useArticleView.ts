import { useMemo } from "react";
import type { Article } from "./types";
import { useArticleUIStore } from "@/store/article-ui-store";

export function useArticleView(article: Article): Article {
  const ui = useArticleUIStore((s) => s.byId[article.id]);

  return useMemo(() => {
    if (!ui) return article;
    return {
      ...article,
      is_bookmarked: ui.is_bookmarked,
      user_like_status: ui.user_like_status,
      likes_count: ui.likes_count,
      dislikes_count: ui.dislikes_count,
    };
  }, [article, ui]);
}
