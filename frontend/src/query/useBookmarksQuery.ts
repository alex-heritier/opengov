import { useQuery } from "@tanstack/react-query";
import client from "@/api/client";
import { BookmarkedArticle } from "@/hook/types";
import { useArticleUIStore } from "@/store/article-ui-store";

export function useBookmarksQuery() {
  const hydrate = useArticleUIStore((s) => s.hydrate);

  return useQuery({
    queryKey: ["bookmarks"],
    queryFn: async () => {
      const { data } = await client.get<{ articles: BookmarkedArticle[] }>(
        "/api/bookmarks",
      );
      hydrate(data.articles);
      return data.articles;
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}
