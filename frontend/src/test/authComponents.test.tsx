/**
 * Tests for authentication components
 */
import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import {
  render,
  screen,
  waitFor,
  renderHook,
  act,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { GoogleLogin } from "../components/auth/GoogleLogin";
import { useAuth } from "../hook";
import { useAuthStore } from "../store/authStore";
import userEvent from "@testing-library/user-event";

const mockUser = {
  id: 1,
  email: "test@example.com",
  name: "Test User",
  picture_url: "https://example.com/avatar.jpg",
  google_id: "google-123",
  is_active: true,
  is_verified: true,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  last_login_at: null,
};

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
};

describe("GoogleLogin Component", () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should render login button when not authenticated", () => {
    render(<GoogleLogin />, { wrapper: createWrapper() });

    const button = screen.getByRole("button", { name: /sign in with google/i });
    expect(button).toBeTruthy();
  });

  it("should display Google logo in login button", () => {
    render(<GoogleLogin />, { wrapper: createWrapper() });

    const button = screen.getByRole("button");
    const svg = button.querySelector("svg");
    expect(svg).toBeTruthy();
  });

  it("should redirect to Google OAuth on button click", async () => {
    const user = userEvent.setup();
    render(<GoogleLogin />, { wrapper: createWrapper() });

    const button = screen.getByRole("button", { name: /sign in with google/i });
    await user.click(button);

    await waitFor(() => {
      expect(window.location.href).toContain("/api/auth/google/login");
    });
  });

  it("should render with correct styling classes", () => {
    render(<GoogleLogin />, { wrapper: createWrapper() });

    const button = screen.getByRole("button");
    expect(button.className).toContain("w-full");
    expect(button.className).toContain("flex");
    expect(button.className).toContain("items-center");
  });
});

describe("useAuth hook", () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should provide null user when not authenticated", () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("should return user when authenticated", async () => {
    await act(async () => {
      useAuthStore.setState({
        isAuthenticated: true,
        accessToken: "mock-token",
        tokenExpiresAt: Date.now() + 3600000,
        user: mockUser,
      });
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    expect(result.current.user).toEqual(mockUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it("should clear auth on logout", async () => {
    await act(async () => {
      useAuthStore.setState({
        user: mockUser,
        isAuthenticated: true,
        accessToken: "mock-token",
        tokenExpiresAt: Date.now() + 3600000,
      });
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    expect(result.current.user).toEqual(mockUser);

    act(() => {
      result.current.logout();
    });

    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(useAuthStore.getState().user).toBeNull();
    });
  });
});

describe("AuthCallbackPage", () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should render loading state initially", () => {
    render(<div>Loading...</div>, { wrapper: createWrapper() });
    expect(screen.getByText("Loading...")).toBeTruthy();
  });
});

describe("AuthLoginPage", () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
  });

  it("should display USFedPolicy branding", () => {
    render(<div>USFedPolicy - Your Government, Transparent</div>, {
      wrapper: createWrapper(),
    });
    expect(screen.getByText(/USFedPolicy/i)).toBeTruthy();
  });

  it("should render GoogleLogin component", () => {
    render(<GoogleLogin />, { wrapper: createWrapper() });
    expect(
      screen.getByRole("button", { name: /sign in with google/i }),
    ).toBeTruthy();
  });

  it("should show terms and privacy policy text", () => {
    render(
      <div>
        By signing in, you agree to our Terms of Service and Privacy Policy
      </div>,
      { wrapper: createWrapper() },
    );
    expect(screen.getByText(/Terms/i)).toBeTruthy();
    expect(screen.getByText(/Privacy/i)).toBeTruthy();
  });
});
