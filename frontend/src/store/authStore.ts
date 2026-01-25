import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { jwtDecode } from "jwt-decode";
import type { User } from "@/hook/types";

interface JWTPayload {
  exp: number;
  [key: string]: any;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  tokenExpiresAt: number | null;
  isAuthenticated: boolean;

  // Actions
  setAuth: (accessToken: string, user: User) => void;
  updateUser: (user: User) => void;
  clearAuth: () => void;
  isTokenExpiringSoon: () => boolean;
}

// Calculate token expiration from JWT
const getTokenExpiration = (token: string): number => {
  try {
    const decoded = jwtDecode<JWTPayload>(token);
    return decoded.exp * 1000; // Convert to milliseconds
  } catch {
    return 0;
  }
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      tokenExpiresAt: null,
      isAuthenticated: false,

      setAuth: (accessToken: string, user: User) => {
        const expiresAt = getTokenExpiration(accessToken);
        set({
          user,
          accessToken,
          tokenExpiresAt: expiresAt,
          isAuthenticated: true,
        });
      },

      updateUser: (user: User) => {
        set({ user });
      },

      clearAuth: () => {
        set({
          user: null,
          accessToken: null,
          tokenExpiresAt: null,
          isAuthenticated: false,
        });
      },

      isTokenExpiringSoon: () => {
        const { tokenExpiresAt } = get();
        if (!tokenExpiresAt) return false;

        const now = Date.now();
        const timeLeft = tokenExpiresAt - now;

        // Consider token expiring soon if less than 10 minutes left
        const tenMinutesInMs = 10 * 60 * 1000;
        return timeLeft < tenMinutesInMs && timeLeft > 0;
      },
    }),
    {
      name: "opengov-auth",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        tokenExpiresAt: state.tokenExpiresAt,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
);
