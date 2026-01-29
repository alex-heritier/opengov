import React from "react";
import {
  ExternalLink,
  FileText,
  Bookmark,
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
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

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
  published_at,
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

  // Format date nicely
  const formattedDate = published_at
    ? new Date(published_at).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
    : "Date unavailable";

  return (
    <article
      className="group relative bg-card border border-border rounded-md shadow-[0_1px_2px_rgba(0,0,0,0.04)] hover:shadow-[0_4px_12px_rgba(0,0,0,0.08)] hover:-translate-y-0.5 transition-all duration-200 ease-out"
      aria-labelledby={`entry-title-${id}`}
    >
      <div className="p-5 space-y-4">
        {/* Meta row - cleaner pill design */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className="font-mono text-[10px] uppercase tracking-wider bg-accent/50 text-accent-foreground border-primary/20"
            >
              Fed Register
            </Badge>
            <time className="text-xs font-mono text-muted-foreground/70 tabular-nums">
              {formattedDate}
            </time>
          </div>

          {/* Subtle bookmark icon - cleaner than button text */}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/5"
            onClick={handleToggleBookmark}
          >
            <Bookmark
              className={cn(
                "w-4 h-4 transition-all",
                bookmarked && "fill-current text-primary",
              )}
            />
          </Button>
        </div>

        {/* Title - better hierarchy with Chicago font */}
        <h3
          id={`entry-title-${id}`}
          className="font-chicago text-lg leading-snug text-foreground group-hover:text-primary transition-colors"
        >
          <Link
            to="/feed/$id"
            params={{ id: String(id) }}
            className="focus:outline-none focus:underline decoration-2 underline-offset-2"
          >
            {title}
          </Link>
        </h3>

        {/* Summary - improved readability */}
        <div
          className="text-sm leading-relaxed text-muted-foreground line-clamp-2 font-sans"
          dangerouslySetInnerHTML={{ __html: sanitizedSummary }}
        />

        {/* Actions - minimal icon row */}
        <div className="flex items-center justify-between pt-3 border-t border-border/50">
          <div className="flex gap-1">
            <Button
              variant="ghost"
              size="sm"
              asChild
              className="h-8 text-xs font-medium text-muted-foreground hover:text-foreground gap-1.5 px-2"
            >
              <Link to="/feed/$id" params={{ id: String(id) }}>
                <FileText className="w-3.5 h-3.5" />
                Read
              </Link>
            </Button>
            {source_url && (
              <Button
                variant="ghost"
                size="sm"
                asChild
                className="h-8 text-xs font-medium text-muted-foreground hover:text-foreground gap-1.5 px-2"
              >
                <a href={source_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="w-3.5 h-3.5" />
                  Source
                </a>
              </Button>
            )}
          </div>

          {/* Voting - cleaner segmented control style */}
          <div className="flex items-center bg-secondary/50 rounded-md p-0.5">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLike}
              className={cn(
                "h-7 px-2 text-xs gap-1 rounded-sm transition-all",
                likeStatus === true
                  ? "bg-white text-primary shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <ThumbsUp className="w-3.5 h-3.5" />
              <span className="font-mono min-w-[1rem] text-center">
                {likesCount}
              </span>
            </Button>
            <div className="w-px h-4 bg-border mx-0.5" />
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDislike}
              className={cn(
                "h-7 px-2 text-xs gap-1 rounded-sm transition-all",
                likeStatus === false
                  ? "bg-white text-destructive shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <ThumbsDown className="w-3.5 h-3.5" />
              <span className="font-mono min-w-[1rem] text-center">
                {dislikesCount}
              </span>
            </Button>
          </div>
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
