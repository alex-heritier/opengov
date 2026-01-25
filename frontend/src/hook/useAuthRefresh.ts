import { useEffect, useRef } from "react";
import axios from "axios";
import { useAuthStore } from "@/store/authStore";

const API_URL =
  (import.meta.env.VITE_API_URL as string) || "http://localhost:8000";

function redirectToLogin(): void {
  if (
    typeof window !== "undefined" &&
    !window.location.pathname.startsWith("/auth")
  ) {
    window.location.href = "/auth/login";
  }
}

export function useAuthRefresh() {
  const isRefreshingRef = useRef(false);

  useEffect(() => {
    const refreshIfNeeded = async () => {
      const { accessToken, isTokenExpiringSoon, setAuth, clearAuth, user } =
        useAuthStore.getState();

      if (!accessToken) return;
      if (!isTokenExpiringSoon()) return;
      if (isRefreshingRef.current) return;

      isRefreshingRef.current = true;
      try {
        const refreshResponse = await axios.post(
          `${API_URL}/api/auth/refresh`,
          {},
          { headers: { Authorization: `Bearer ${accessToken}` } },
        );

        const newToken: string | undefined = refreshResponse.data?.access_token;
        if (!newToken) {
          clearAuth();
          redirectToLogin();
          return;
        }

        if (user) {
          setAuth(newToken, user);
          return;
        }

        const meResponse = await axios.get(`${API_URL}/api/auth/me`, {
          headers: { Authorization: `Bearer ${newToken}` },
        });

        setAuth(newToken, meResponse.data);
      } catch (error) {
        console.error("Token refresh failed:", error);
        clearAuth();
        redirectToLogin();
      } finally {
        isRefreshingRef.current = false;
      }
    };

    // Run once on mount, then periodically.
    refreshIfNeeded();
    const interval = setInterval(refreshIfNeeded, 60_000);
    return () => clearInterval(interval);
  }, []);
}
