import { useEffect, useCallback, useMemo } from "react";
import { useFeedQuery, useFeedStore, type FeedEntryResponse } from "@/hook";
import type { FeedResponse } from "@/query";
import { FeedEntryCard } from "./FeedEntryCard";
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

  const items = useMemo<FeedEntryResponse[]>(() => {
    const allItems =
      data?.pages.flatMap((page: FeedResponse) => page.items) ?? [];
    const seen = new Set<number>();
    return allItems.filter((item) => {
      if (seen.has(item.id)) return false;
      seen.add(item.id);
      return true;
    });
  }, [data]);

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
      <Alert
        variant="destructive"
        className="rounded-lg border border-destructive/20 bg-destructive/5 text-destructive"
      >
        <AlertCircle className="h-4 w-4" />
        <AlertDescription className="font-medium">
          Failed to load feed entries. Please try again later.
        </AlertDescription>
      </Alert>
    );
  }

  const showEmptyState = items.length === 0 && !isLoading;
  const showLoadingMore = isFetchingNextPage;

  return (
    <div className="space-y-4">
      <div className="">
        {items.map((item) => (
          <FeedEntryCard
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
        <div className="text-center py-16 px-4">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/5 mb-4">
            <AlertCircle className="w-8 h-8 text-primary/40" />
          </div>
          <h3 className="font-chicago text-lg text-foreground mb-2">
            No documents found
          </h3>
          <p className="text-muted-foreground text-sm max-w-sm mx-auto leading-relaxed">
            Looks like the federal register is quiet today, or try adjusting
            your search terms.
          </p>
        </div>
      )}

      {(isLoading || showLoadingMore) && (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div
              key={i}
              className="bg-card border border-border rounded-lg p-6 shadow-sm space-y-4"
            >
              <div className="flex gap-3 mb-2">
                <Skeleton className="h-4 w-24 rounded bg-secondary" />
                <Skeleton className="h-4 w-32 rounded bg-secondary" />
              </div>
              <Skeleton className="h-8 w-3/4 rounded bg-secondary" />
              <div className="space-y-2">
                <Skeleton className="h-4 w-full rounded bg-secondary" />
                <Skeleton className="h-4 w-5/6 rounded bg-secondary" />
                <Skeleton className="h-4 w-4/6 rounded bg-secondary" />
              </div>
              <div className="flex gap-4 pt-4 border-t border-border mt-2">
                <Skeleton className="h-8 w-24 rounded bg-secondary" />
                <Skeleton className="h-8 w-24 rounded bg-secondary" />
                <Skeleton className="h-8 w-24 rounded bg-secondary" />
              </div>
            </div>
          ))}
        </div>
      )}

      {!hasNextPage && items.length > 0 && !isLoading && (
        <div className="text-center py-12 border-t border-border border-dashed mt-8">
          <p className="text-sm font-medium text-muted-foreground">
            End of Feed
          </p>
        </div>
      )}
    </div>
  );
};
