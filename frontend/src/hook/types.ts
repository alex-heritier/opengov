export interface User {
  id: number
  email: string
  name: string | null
  picture_url: string | null
  google_id: string | null
  political_leaning: string | null
  is_active: boolean
  is_verified: boolean
  created_at: string
  updated_at: string
  last_login_at: string | null
}

export interface Article {
  id: number
  title: string
  summary: string
  source_url: string
  published_at: string
  created_at: string
  is_bookmarked?: boolean
  document_number?: string
  unique_key?: string
  user_like_status?: boolean | null  // null = no vote, true = liked, false = disliked
  likes_count?: number
  dislikes_count?: number
}

export type BookmarkedArticle = Article & {
  is_bookmarked: true
  document_number: string
  unique_key: string
  updated_at: string
  likes_count: number
  dislikes_count: number
}
