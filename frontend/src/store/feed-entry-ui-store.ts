import { createWithEqualityFn } from "zustand/traditional";

export type LikeStatus = boolean | null; // null=no vote, true=like, false=dislike

export interface FeedEntryUIState {
  is_bookmarked: boolean;
  user_like_status: LikeStatus;
  likes_count: number;
  dislikes_count: number;
}

interface FeedEntryUIStore {
  byId: Record<number, FeedEntryUIState>;

  hydrate: (
    entries: Array<{
      id: number;
      is_bookmarked?: boolean;
      user_like_status?: number | null;
      likes_count?: number;
      dislikes_count?: number;
    }>,
  ) => void;

  setBookmark: (id: number, value: boolean) => void;
  setLikeStatus: (id: number, value: LikeStatus) => void;

  applyReaction: (id: number, next: LikeStatus) => { prev: FeedEntryUIState };
  restore: (id: number, prev: FeedEntryUIState) => void;
}

const defaultUI = (): FeedEntryUIState => ({
  is_bookmarked: false,
  user_like_status: null,
  likes_count: 0,
  dislikes_count: 0,
});

function convertLikeStatus(status: number | null | undefined): LikeStatus {
  if (status === undefined || status === null) return null;
  if (status === 1) return true;
  if (status === -1) return false;
  return null;
}

export const useFeedEntryUIStore = createWithEqualityFn<FeedEntryUIStore>((set, get) => ({
  byId: {},

  hydrate: (entries) => {
    set((s) => {
      const next = { ...s.byId };
      for (const entry of entries) {
        const existing = next[entry.id] ?? defaultUI();
        next[entry.id] = {
          is_bookmarked: entry.is_bookmarked ?? existing.is_bookmarked,
          user_like_status:
            entry.user_like_status === undefined
              ? existing.user_like_status
              : convertLikeStatus(entry.user_like_status),
          likes_count: entry.likes_count ?? existing.likes_count,
          dislikes_count: entry.dislikes_count ?? existing.dislikes_count,
        };
      }
      return { byId: next };
    });
  },

  setBookmark: (id, value) =>
    set((s) => ({
      byId: {
        ...s.byId,
        [id]: { ...(s.byId[id] ?? defaultUI()), is_bookmarked: value },
      },
    })),

  setLikeStatus: (id, value) =>
    set((s) => ({
      byId: {
        ...s.byId,
        [id]: { ...(s.byId[id] ?? defaultUI()), user_like_status: value },
      },
    })),

  applyReaction: (id, next) => {
    const prev = get().byId[id] ?? defaultUI();

    let likes = prev.likes_count;
    let dislikes = prev.dislikes_count;

    // Remove previous vote
    if (prev.user_like_status === true) likes -= 1;
    if (prev.user_like_status === false) dislikes -= 1;

    // Add new vote
    if (next === true) likes += 1;
    if (next === false) dislikes += 1;

    const updated: FeedEntryUIState = {
      ...prev,
      user_like_status: next,
      likes_count: Math.max(0, likes),
      dislikes_count: Math.max(0, dislikes),
    };

    set((s) => ({ byId: { ...s.byId, [id]: updated } }));
    return { prev };
  },

  restore: (id, prev) => set((s) => ({ byId: { ...s.byId, [id]: prev } })),
}));
