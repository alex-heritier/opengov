import { describe, it, expect, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { FeedEntryCard } from "./FeedEntryCard";
import { useFeedEntryUIStore } from "@/store/feed-entry-ui-store";

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    to,
    children,
    ...props
  }: {
    to: string;
    children: React.ReactNode;
  }) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  useNavigate: () => vi.fn(),
}));

vi.mock("@/hook", () => {
  const base = { mutate: vi.fn(), isPending: false };
  return {
    useToggleBookmarkMutation: () => base,
    useToggleLikeMutation: () => base,
    useRemoveLikeMutation: () => base,
    useAuth: () => ({ isAuthenticated: true }),
  };
});

// Create a test QueryClient
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

describe("FeedEntryCard", () => {
  const queryClient = createTestQueryClient();

  const renderWithProviders = (component: React.ReactNode) => {
    return render(
      <QueryClientProvider client={queryClient}>
        {component}
      </QueryClientProvider>,
    );
  };

  it("renders feed entry title", () => {
    renderWithProviders(
      <FeedEntryCard
        id={1}
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />,
    );

    expect(screen.getByText("Test Article")).toBeInTheDocument();
  });

  it("renders feed entry title as link when id is present", () => {
    renderWithProviders(
      <FeedEntryCard
        id={123}
        title="Test Article with Link"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />,
    );

    const link = screen.getByRole("link", { name: "Test Article with Link" });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", "/feed/$id");
  });

  it("renders feed entry summary", () => {
    renderWithProviders(
      <FeedEntryCard
        id={1}
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />,
    );

    expect(screen.getByText("Test summary")).toBeInTheDocument();
  });

  it("renders source link", () => {
    renderWithProviders(
      <FeedEntryCard
        id={1}
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />,
    );

    const link = screen.getByRole("link", { name: /Source/i });
    expect(link).toHaveAttribute("href", "https://example.com");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("does not fall back to props when store likeStatus is null", () => {
    const id = 123;

    act(() => {
      useFeedEntryUIStore.setState({
        byId: {
          [id]: {
            is_bookmarked: false,
            user_like_status: null,
            likes_count: 0,
            dislikes_count: 0,
          },
        },
      });
    });

    renderWithProviders(
      <FeedEntryCard
        id={id}
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
        user_like_status={-1}
      />,
    );

    // The thumbs down button should not be highlighted since store has null
    const buttons = screen.getAllByRole("button");
    const thumbsDownBtn = buttons.find((btn) => btn.textContent === "0");
    expect(thumbsDownBtn).toBeDefined();
    expect(thumbsDownBtn).not.toHaveClass("bg-red-100");

    act(() => {
      useFeedEntryUIStore.setState({ byId: {} });
    });
  });
});
