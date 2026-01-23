import { create } from 'zustand'

export type LikeStatus = boolean | null // null=no vote, true=like, false=dislike

export interface ArticleUIState {
  is_bookmarked: boolean
  user_like_status: LikeStatus
  likes_count: number
  dislikes_count: number
}

interface ArticleUIStore {
  byId: Record<number, ArticleUIState>

  hydrate: (articles: Array<{
    id: number
    is_bookmarked?: boolean
    user_like_status?: boolean | null
    likes_count?: number
    dislikes_count?: number
  }>) => void

  setBookmark: (id: number, value: boolean) => void
  setLikeStatus: (id: number, value: LikeStatus) => void

  applyReaction: (id: number, next: LikeStatus) => { prev: ArticleUIState }
  restore: (id: number, prev: ArticleUIState) => void
}

const defaultUI = (): ArticleUIState => ({
  is_bookmarked: false,
  user_like_status: null,
  likes_count: 0,
  dislikes_count: 0,
})

export const useArticleUIStore = create<ArticleUIStore>((set, get) => ({
  byId: {},

  hydrate: (articles) => {
    set((s) => {
      const next = { ...s.byId }
      for (const a of articles) {
        const existing = next[a.id] ?? defaultUI()
        next[a.id] = {
          is_bookmarked: a.is_bookmarked ?? existing.is_bookmarked,
          user_like_status:
            a.user_like_status === undefined ? existing.user_like_status : a.user_like_status,
          likes_count: a.likes_count ?? existing.likes_count,
          dislikes_count: a.dislikes_count ?? existing.dislikes_count,
        }
      }
      return { byId: next }
    })
  },

  setBookmark: (id, value) =>
    set((s) => ({
      byId: { ...s.byId, [id]: { ...(s.byId[id] ?? defaultUI()), is_bookmarked: value } },
    })),

  setLikeStatus: (id, value) =>
    set((s) => ({
      byId: { ...s.byId, [id]: { ...(s.byId[id] ?? defaultUI()), user_like_status: value } },
    })),

  applyReaction: (id, next) => {
    const prev = get().byId[id] ?? defaultUI()

    let likes = prev.likes_count
    let dislikes = prev.dislikes_count

    // Remove previous vote
    if (prev.user_like_status === true) likes -= 1
    if (prev.user_like_status === false) dislikes -= 1

    // Add new vote
    if (next === true) likes += 1
    if (next === false) dislikes += 1

    const updated: ArticleUIState = {
      ...prev,
      user_like_status: next,
      likes_count: Math.max(0, likes),
      dislikes_count: Math.max(0, dislikes),
    }

    set((s) => ({ byId: { ...s.byId, [id]: updated } }))
    return { prev }
  },

  restore: (id, prev) => set((s) => ({ byId: { ...s.byId, [id]: prev } })),
}))
