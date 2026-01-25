export interface User {
  id: number;
  email: string;
  name: string | null;
  picture_url: string | null;
  google_id: string | null;
  political_leaning: string | null;
  state: string | null;
  is_active: boolean;
  is_verified: boolean;
  created_at: string;
  updated_at: string;
  last_login_at: string | null;
}

export interface Article {
  id: number;
  title: string;
  summary: string;
  keypoints?: string[];
  impact_score?: "low" | "medium" | "high" | null;
  political_score?: number | null;
  source_url: string;
  published_at: string;
  created_at: string;
  is_bookmarked?: boolean;
  user_like_status?: number | null;
  likes_count?: number;
  dislikes_count?: number;
}

export interface FeedEntryResponse {
  id: number;
  title: string;
  summary: string;
  keypoints?: string[];
  impact_score?: string | null;
  political_score?: number | null;
  source_url: string;
  published_at: string;
  is_bookmarked?: boolean;
  user_like_status?: number | null;
  likes_count: number;
  dislikes_count: number;
}

export type BookmarkedArticle = FeedEntryResponse & {
  is_bookmarked: true;
  updated_at: string;
};

export interface FeedResponse {
  items: FeedEntryResponse[];
  page: number;
  limit: number;
  total: number;
  has_next: boolean;
}
