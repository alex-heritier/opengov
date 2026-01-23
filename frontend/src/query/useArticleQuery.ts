import { useQuery } from "@tanstack/react-query";
import client from "@/api/client";
import { Article } from "@/hook/types";
import { useArticleUIStore } from "@/store/article-ui-store";

export interface ArticleDetail extends Article {
  updated_at: string;
}

export function useArticleQuery(id: number) {
  const hydrate = useArticleUIStore((s) => s.hydrate);

  return useQuery({
    queryKey: ["article", id],
    queryFn: async () => {
      const { data } = await client.get<Article>(`/api/feed/${id}`);
      hydrate([data]);
      return data;
    },
    enabled: !!id,
    staleTime: 1000 * 60 * 10,
  });
}

export function useArticleBySlugQuery(slug: string) {
  const hydrate = useArticleUIStore((s) => s.hydrate);

  return useQuery({
    queryKey: ["article", "slug", slug],
    queryFn: async () => {
      const { data } = await client.get<ArticleDetail>(
        `/api/feed/slug/${slug}`,
      );
      hydrate([{ ...data, id: data.id }]);
      return data;
    },
    enabled: !!slug,
    staleTime: 1000 * 60 * 10,
  });
}
