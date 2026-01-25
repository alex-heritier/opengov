import React from "react";
import {
  ExternalLink,
  FileText,
  Bookmark,
  BookmarkCheck,
  ThumbsUp,
  ThumbsDown,
} from "lucide-react";
import { Link, useNavigate } from "@tanstack/react-router";
import DOMPurify from "dompurify";
import {
  useToggleBookmarkMutation,
  useToggleLikeMutation,
  useRemoveLikeMutation,
  useAuth,
} from "@/hook";
import { useArticleUIStore, type LikeStatus } from "@/store/article-ui-store";

interface ArticleCardProps {
  id?: number;
  title: string;
  summary: string;
  source_url: string;
  published_at: string;
  unique_key?: string | null;
  is_bookmarked?: boolean;
  user_like_status?: boolean | null;
  likes_count?: number;
  dislikes_count?: number;
}

export const ArticleCard: React.FC<ArticleCardProps> = ({
  id,
  title,
  summary,
  source_url,
  published_at: _published_at,
  unique_key,
  is_bookmarked = false,
  user_like_status = null,
  likes_count = 0,
  dislikes_count = 0,
}) => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const toggleBookmark = useToggleBookmarkMutation();
  const toggleLike = useToggleLikeMutation();
  const removeLike = useRemoveLikeMutation();

  const ui = useArticleUIStore((s) => (id ? s.byId[id] : undefined));

  const bookmarked = ui?.is_bookmarked ?? is_bookmarked;
  // user_like_status is tri-state (true/false/null). Only fall back to props
  // when the store has no value (undefined), not when it's explicitly null.
  const likeStatus: LikeStatus =
    ui?.user_like_status === undefined ? user_like_status : ui.user_like_status;
  const likesCount = ui?.likes_count ?? likes_count;
  const dislikesCount = ui?.dislikes_count ?? dislikes_count;

  const requireAuth = () => {
    if (!isAuthenticated) {
      navigate({ to: "/login" });
      return false;
    }
    return true;
  };

  const sanitizedSummary = DOMPurify.sanitize(summary, {
    ALLOWED_TAGS: ["b", "i", "em", "strong", "br", "p"],
    ALLOWED_ATTR: [],
  });

  const handleToggleBookmark = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!requireAuth() || !id) return;
    toggleBookmark.mutate(id);
  };

  const handleLike = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!requireAuth() || !id) return;
    if (likeStatus === true) {
      removeLike.mutate(id);
    } else {
      toggleLike.mutate({ articleId: id, isPositive: true });
    }
  };

  const handleDislike = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!requireAuth() || !id) return;
    if (likeStatus === false) {
      removeLike.mutate(id);
    } else {
      toggleLike.mutate({ articleId: id, isPositive: false });
    }
  };

  return (
    <article className="border-b border-gray-200 py-4 sm:py-6 hover:bg-gray-50 transition-colors">
      <div className="space-y-3">
        <h3 className="text-base sm:text-lg font-bold text-gray-900 leading-snug">
          {unique_key ? (
            <Link
              to="/articles/$slug"
              params={{ slug: unique_key }}
              className="hover:underline hover:text-blue-700 transition-colors"
            >
              {title}
            </Link>
          ) : (
            title
          )}
        </h3>
        <p
          className="text-sm text-gray-600 line-clamp-3 leading-relaxed"
          dangerouslySetInnerHTML={{ __html: sanitizedSummary }}
        />

        <div className="flex flex-wrap gap-2 pt-2">
          {unique_key && (
            <Link
              to="/articles/$slug"
              params={{ slug: unique_key }}
              className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium bg-gray-100 hover:bg-gray-200 transition-colors text-gray-900 no-underline"
            >
              <FileText className="w-4 h-4" />
              View Details
            </Link>
          )}
          <a
            href={source_url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Read on Federal Register"
            className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium border border-gray-300 bg-white hover:bg-gray-50 transition-colors cursor-pointer text-gray-900 no-underline"
          >
            <ExternalLink className="w-4 h-4" />
            Federal Register
          </a>
          {isAuthenticated && (
            <>
              <button
                onClick={handleLike}
                disabled={toggleLike.isPending || removeLike.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  likeStatus === true
                    ? "bg-green-600 text-white"
                    : "border border-gray-300 bg-white hover:bg-gray-50"
                }`}
                aria-label="Like article"
              >
                <ThumbsUp className="w-4 h-4" />
                {likesCount > 0 && <span>{likesCount}</span>}
              </button>
              <button
                onClick={handleDislike}
                disabled={toggleLike.isPending || removeLike.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  likeStatus === false
                    ? "bg-red-600 text-white"
                    : "border border-gray-300 bg-white hover:bg-gray-50"
                }`}
                aria-label="Dislike article"
              >
                <ThumbsDown className="w-4 h-4" />
                {dislikesCount > 0 && <span>{dislikesCount}</span>}
              </button>
              <button
                onClick={handleToggleBookmark}
                disabled={toggleBookmark.isPending}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors ${
                  bookmarked
                    ? "bg-blue-600 text-white"
                    : "border border-gray-300 bg-white hover:bg-gray-50"
                }`}
                aria-label={bookmarked ? "Remove bookmark" : "Bookmark article"}
              >
                {bookmarked ? (
                  <>
                    <BookmarkCheck className="w-4 h-4" />
                    Bookmarked
                  </>
                ) : (
                  <>
                    <Bookmark className="w-4 h-4" />
                    Bookmark
                  </>
                )}
              </button>
            </>
          )}
        </div>
      </div>
    </article>
  );
};
