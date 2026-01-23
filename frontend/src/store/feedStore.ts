import { create } from "zustand";

interface FeedStore {
  sort: "newest" | "oldest";
  pageSize: number;
  setSortOrder: (sort: "newest" | "oldest") => void;
  setPageSize: (size: number) => void;
}

export const useFeedStore = create<FeedStore>((set) => ({
  sort: "newest",
  pageSize: 20,
  setSortOrder: (sort) => set({ sort }),
  setPageSize: (pageSize) => set({ pageSize }),
}));
