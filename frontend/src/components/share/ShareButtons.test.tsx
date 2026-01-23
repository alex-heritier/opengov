import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";
import ShareButtons from "./ShareButtons";

describe("ShareButtons", () => {
  const mockProps = {
    title: "Test Article Title",
    url: "https://example.com/article/123",
    summary: "This is a test summary",
  };

  beforeEach(() => {
    // Mock clipboard API
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn(() => Promise.resolve()),
      },
    });
  });

  it("renders all share buttons", () => {
    render(<ShareButtons {...mockProps} />);

    expect(screen.getByLabelText("Share on Twitter")).toBeInTheDocument();
    expect(screen.getByLabelText("Share on Facebook")).toBeInTheDocument();
    expect(screen.getByLabelText("Share on LinkedIn")).toBeInTheDocument();
    expect(screen.getByLabelText("Share via Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Copy link to clipboard")).toBeInTheDocument();
  });

  it("generates correct Twitter share URL", () => {
    render(<ShareButtons {...mockProps} />);

    const twitterLink = screen.getByLabelText(
      "Share on Twitter",
    ) as HTMLAnchorElement;
    expect(twitterLink.href).toContain("twitter.com/intent/tweet");
    expect(twitterLink.href).toContain(encodeURIComponent(mockProps.title));
    expect(twitterLink.href).toContain(encodeURIComponent(mockProps.url));
  });

  it("generates correct Facebook share URL", () => {
    render(<ShareButtons {...mockProps} />);

    const facebookLink = screen.getByLabelText(
      "Share on Facebook",
    ) as HTMLAnchorElement;
    expect(facebookLink.href).toContain("facebook.com/sharer");
    expect(facebookLink.href).toContain(encodeURIComponent(mockProps.url));
  });

  it("generates correct LinkedIn share URL", () => {
    render(<ShareButtons {...mockProps} />);

    const linkedinLink = screen.getByLabelText(
      "Share on LinkedIn",
    ) as HTMLAnchorElement;
    expect(linkedinLink.href).toContain("linkedin.com/sharing");
    expect(linkedinLink.href).toContain(encodeURIComponent(mockProps.url));
  });

  it("generates correct email share link", () => {
    render(<ShareButtons {...mockProps} />);

    const emailLink = screen.getByLabelText(
      "Share via Email",
    ) as HTMLAnchorElement;
    expect(emailLink.href).toContain("mailto:");
    expect(emailLink.href).toContain(encodeURIComponent(mockProps.title));
    expect(emailLink.href).toContain(encodeURIComponent(mockProps.url));
  });

  it("copies link to clipboard when copy button is clicked", async () => {
    render(<ShareButtons {...mockProps} />);

    const copyButton = screen.getByLabelText("Copy link to clipboard");
    fireEvent.click(copyButton);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(mockProps.url);

    await waitFor(() => {
      expect(screen.getByText("Copied!")).toBeInTheDocument();
    });
  });

  it("renders without summary prop", () => {
    const propsWithoutSummary = {
      title: mockProps.title,
      url: mockProps.url,
    };

    render(<ShareButtons {...propsWithoutSummary} />);

    expect(screen.getByLabelText("Share on Twitter")).toBeInTheDocument();
    expect(screen.getByLabelText("Share on Facebook")).toBeInTheDocument();
  });
});
