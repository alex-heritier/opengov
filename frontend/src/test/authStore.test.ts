/**
 * Tests for authentication Zustand store
 */
import { describe, it, expect, beforeEach } from "vitest";
import { useAuthStore } from "../store/authStore";

describe("useAuthStore", () => {
  beforeEach(() => {
    // Clear store before each test
    useAuthStore.getState().clearAuth();
  });

  describe("Initial State", () => {
    it("should have null values initially", () => {
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.tokenExpiresAt).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe("setAuth", () => {
    it("should set authentication state correctly", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: "https://example.com/pic.jpg",
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: "2025-01-01T00:00:00Z",
      };

      // Create a mock JWT token with 1 hour expiration
      const now = Math.floor(Date.now() / 1000);
      const exp = now + 3600; // 1 hour from now
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      const { setAuth } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      const state = useAuthStore.getState();
      expect(state.user).toEqual(mockUser);
      expect(state.accessToken).toBe(mockToken);
      expect(state.isAuthenticated).toBe(true);
      expect(state.tokenExpiresAt).toBeGreaterThan(Date.now());
    });

    it("should extract expiration time from JWT token", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: null,
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: null,
      };

      const now = Math.floor(Date.now() / 1000);
      const exp = now + 7200; // 2 hours from now
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      const { setAuth } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      const state = useAuthStore.getState();
      const expectedExpiration = exp * 1000; // Convert to milliseconds
      expect(state.tokenExpiresAt).toBe(expectedExpiration);
    });
  });

  describe("clearAuth", () => {
    it("should clear all authentication state", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: null,
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: null,
      };

      const now = Math.floor(Date.now() / 1000);
      const exp = now + 3600;
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      // Set auth first
      const { setAuth, clearAuth } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      // Verify it's set
      let state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);

      // Clear auth
      clearAuth();

      // Verify it's cleared
      state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.tokenExpiresAt).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe("isTokenExpiringSoon", () => {
    it("should return false when no token exists", () => {
      const { isTokenExpiringSoon } = useAuthStore.getState();
      expect(isTokenExpiringSoon()).toBe(false);
    });

    it("should return false when token has more than 10 minutes left", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: null,
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: null,
      };

      // Token expires in 30 minutes
      const now = Math.floor(Date.now() / 1000);
      const exp = now + 1800; // 30 minutes
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      const { setAuth, isTokenExpiringSoon } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      expect(isTokenExpiringSoon()).toBe(false);
    });

    it("should return true when token has less than 10 minutes left", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: null,
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: null,
      };

      // Token expires in 5 minutes
      const now = Math.floor(Date.now() / 1000);
      const exp = now + 300; // 5 minutes
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      const { setAuth, isTokenExpiringSoon } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      expect(isTokenExpiringSoon()).toBe(true);
    });

    it("should return false when token is already expired", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        picture_url: null,
        google_id: "google123",
        is_active: true,
        is_verified: true,
        created_at: "2025-01-01T00:00:00Z",
        updated_at: "2025-01-01T00:00:00Z",
        last_login_at: null,
      };

      // Token expired 5 minutes ago
      const now = Math.floor(Date.now() / 1000);
      const exp = now - 300; // 5 minutes ago
      const mockPayload = { sub: 1, exp };
      const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`;

      const { setAuth, isTokenExpiringSoon } = useAuthStore.getState();
      setAuth(mockToken, mockUser);

      expect(isTokenExpiringSoon()).toBe(false);
    });
  });
});
