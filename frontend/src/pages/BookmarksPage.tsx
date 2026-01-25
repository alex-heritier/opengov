import { useBookmarks, useAuth } from "@/hook";
import type { BookmarkedArticle } from "@/hook/types";
import { ArticleCard } from "@/components/feed/ArticleCard";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, Bookmark } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";

export default function BookmarksPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const { bookmarks, isLoading: isBookmarksLoading, error } = useBookmarks();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate({ to: "/auth/login" });
    }
  }, [isAuthenticated, navigate]);

  if (!isAuthenticated) {
    return null; // Will redirect
  }

  if (error) {
    return (
      <div className="w-full max-w-5xl mx-auto px-4 sm:px-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load bookmarks. Please try again later.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (isBookmarksLoading) {
    return (
      <div className="w-full max-w-5xl mx-auto px-4 sm:px-6">
        <div className="space-y-4 sm:space-y-6">
          <div className="flex items-center gap-2">
            <Bookmark className="w-6 h-6 sm:w-8 sm:h-8" />
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900">
              My Bookmarks
            </h1>
          </div>

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
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-5xl mx-auto px-4 sm:px-6">
      <div className="space-y-4 sm:space-y-6">
        {/* Header */}
        <div className="flex items-center gap-2">
          <Bookmark className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900">
            My Bookmarks
          </h1>
        </div>

        {/* Bookmarks List */}
        {bookmarks && bookmarks.length > 0 ? (
          <div className="divide-y divide-gray-200 border-t border-gray-200">
            {bookmarks.map((article: BookmarkedArticle) => (
              <ArticleCard
                key={article.id}
                id={article.id}
                title={article.title}
                summary={article.summary}
                source_url={article.source_url}
                published_at={article.published_at}
                unique_key={article.unique_key}
                is_bookmarked={true}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-12 border border-gray-200 rounded-lg bg-gray-50">
            <Bookmark className="w-12 h-12 mx-auto text-gray-400 mb-4" />
            <p className="text-gray-500 text-lg mb-2">No bookmarks yet</p>
            <p className="text-gray-400 text-sm">
              Start bookmarking articles from the feed to save them for later
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
