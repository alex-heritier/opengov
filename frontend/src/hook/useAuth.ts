/**
 * Auth hook - thin wrapper around authStore for components/pages
 * Provides a stable API while hiding store implementation details
 */
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/store/authStore";
import { shallow } from "zustand/shallow";

export function useAuth() {
  const queryClient = useQueryClient();
  const { user, isAuthenticated, setAuth, updateUser, clearAuth } =
    useAuthStore(
      (s) => ({
        user: s.user,
        isAuthenticated: s.isAuthenticated,
        setAuth: s.setAuth,
        updateUser: s.updateUser,
        clearAuth: s.clearAuth,
      }),
      shallow,
    );

  const login = () => {
    const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8000";
    window.location.href = `${apiUrl}/api/auth/google/login`;
  };

  const logout = () => {
    clearAuth();
    queryClient.clear();
    window.location.href = "/auth/login";
  };

  return {
    user,
    isAuthenticated,
    login,
    logout,
    setAuth,
    updateUser,
  };
}
