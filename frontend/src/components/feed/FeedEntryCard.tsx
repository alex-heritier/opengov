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
import {
  useFeedEntryUIStore,
  type LikeStatus,
} from "@/store/feed-entry-ui-store";
import { useStoreWithEqualityFn } from "zustand/traditional";

interface FeedEntryCardProps {
  id: number;
  title: string;
  summary: string;
  source_url: string;
  published_at: string;
  is_bookmarked?: boolean;
  user_like_status?: number | null;
  likes_count?: number;
  dislikes_count?: number;
}

export const FeedEntryCard: React.FC<FeedEntryCardProps> = ({
  id,
  title,
  summary,
  source_url,
  published_at: _published_at,
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

  const ui = useStoreWithEqualityFn(useFeedEntryUIStore, (s) =>
    id ? s.byId[id] : undefined,
  );

  const bookmarked = ui?.is_bookmarked ?? is_bookmarked;
  const likeStatus: LikeStatus =
    ui?.user_like_status === undefined
      ? convertLikeStatus(user_like_status)
      : ui.user_like_status;
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
      toggleLike.mutate({ feedEntryId: id, isPositive: true });
    }
  };

  const handleDislike = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!requireAuth() || !id) return;
    if (likeStatus === false) {
      removeLike.mutate(id);
    } else {
      toggleLike.mutate({ feedEntryId: id, isPositive: false });
    }
  };

  return (
    <article className="border-b border-gray-200 py-4 sm:py-6 hover:bg-gray-50 transition-colors">
      <div className="space-y-3">
        <h3 className="text-base sm:text-lg font-bold text-gray-900 leading-snug">
          <Link
            to="/feed/$id"
            params={{ id: String(id) }}
            className="hover:underline hover:text-blue-700 transition-colors"
          >
            {title}
          </Link>
        </h3>
        <p
          className="text-sm text-gray-600 line-clamp-3 leading-relaxed"
          dangerouslySetInnerHTML={{ __html: sanitizedSummary }}
        />

        <div className="flex flex-wrap gap-2 pt-2">
          <Link
            to="/feed/$id"
            params={{ id: String(id) }}
            className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium bg-gray-100 hover:bg-gray-200 transition-colors text-gray-900 no-underline"
          >
            <FileText className="w-3.5 h-3.5" />
            <span>View Details</span>
          </Link>
          {source_url && (
            <a
              href={source_url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium bg-blue-50 hover:bg-blue-100 transition-colors text-blue-700 no-underline"
            >
              <ExternalLink className="w-3.5 h-3.5" />
              <span>Source</span>
            </a>
          )}
          <button
            onClick={handleToggleBookmark}
            className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs sm:text-sm font-medium transition-colors no-underline ${
              bookmarked
                ? "bg-blue-600 text-white hover:bg-blue-700"
                : "bg-gray-100 hover:bg-gray-200 text-gray-900"
            }`}
          >
            {bookmarked ? (
              <BookmarkCheck className="w-3.5 h-3.5" />
            ) : (
              <Bookmark className="w-3.5 h-3.5" />
            )}
            <span>{bookmarked ? "Saved" : "Save"}</span>
          </button>
        </div>

        <div className="flex items-center gap-4 text-xs text-gray-500 pt-1">
          <button
            onClick={handleLike}
            className={`flex items-center gap-1.5 px-2 py-1 rounded-md transition-colors ${
              likeStatus === true
                ? "bg-green-100 text-green-700"
                : "hover:bg-gray-100"
            }`}
          >
            <ThumbsUp className="w-3.5 h-3.5" />
            <span>{likesCount}</span>
          </button>
          <button
            onClick={handleDislike}
            className={`flex items-center gap-1.5 px-2 py-1 rounded-md transition-colors ${
              likeStatus === false
                ? "bg-red-100 text-red-700"
                : "hover:bg-gray-100"
            }`}
          >
            <ThumbsDown className="w-3.5 h-3.5" />
            <span>{dislikesCount}</span>
          </button>
        </div>
      </div>
    </article>
  );
};

function convertLikeStatus(status: number | null | undefined): LikeStatus {
  if (status === undefined || status === null) return null;
  if (status === 1) return true;
  if (status === -1) return false;
  return null;
}
