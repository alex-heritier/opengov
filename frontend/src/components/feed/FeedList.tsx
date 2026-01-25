import { useEffect, useCallback, useMemo } from "react";
import { useFeedQuery, useFeedStore, type FeedEntryResponse } from "@/hook";
import type { FeedResponse } from "@/query";
import { ArticleCard } from "./ArticleCard";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";

export const FeedList: React.FC = () => {
  const { sort, pageSize } = useFeedStore();
  const {
    data,
    isLoading,
    isFetchingNextPage,
    error,
    hasNextPage,
    fetchNextPage,
  } = useFeedQuery(pageSize, sort);

  const items = useMemo<FeedEntryResponse[]>(
    () => data?.pages.flatMap((page: FeedResponse) => page.items) ?? [],
    [data],
  );

  const handleScroll = useCallback(() => {
    if (isFetchingNextPage || !hasNextPage) return;

    const scrollHeight = document.documentElement.scrollHeight;
    const scrollTop = document.documentElement.scrollTop;
    const clientHeight = window.innerHeight;

    if (scrollTop + clientHeight >= scrollHeight - 300) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load articles. Please try again later.
        </AlertDescription>
      </Alert>
    );
  }

  const showEmptyState = items.length === 0 && !isLoading;
  const showLoadingMore = isFetchingNextPage;

  return (
    <div className="space-y-0">
      <div className="divide-y divide-gray-200 border-t border-gray-200">
        {items.map((item) => (
          <ArticleCard
            key={item.id}
            id={item.id}
            title={item.title}
            summary={item.summary}
            source_url={item.source_url}
            published_at={item.published_at}
            is_bookmarked={item.is_bookmarked}
            user_like_status={item.user_like_status}
            likes_count={item.likes_count}
            dislikes_count={item.dislikes_count}
          />
        ))}
      </div>

      {showEmptyState && (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg">No articles found.</p>
        </div>
      )}

      {(isLoading || showLoadingMore) && (
        <div className="divide-y divide-gray-200 border-t border-gray-200">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="py-4 sm:py-6 space-y-3">
              <Skeleton className="h-6 w-3/4" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-5/6" />
              <div className="flex gap-2 pt-2">
                <Skeleton className="h-8 w-24" />
                <Skeleton className="h-8 w-32" />
              </div>
            </div>
          ))}
        </div>
      )}

      {!hasNextPage && items.length > 0 && !isLoading && (
        <div className="text-center py-8 border-t border-gray-200">
          <p className="text-sm text-gray-500">No more articles to load.</p>
        </div>
      )}
    </div>
  );
};
