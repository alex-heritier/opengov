import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ArticleCard } from './ArticleCard'

// Create a test QueryClient
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

describe('ArticleCard', () => {
  const queryClient = createTestQueryClient()

  const renderWithProviders = (component: React.ReactNode) => {
    return render(
      <QueryClientProvider client={queryClient}>
        {component}
      </QueryClientProvider>
    )
  }

  it('renders article title', () => {
    renderWithProviders(
      <ArticleCard
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />
    )

    expect(screen.getByText('Test Article')).toBeInTheDocument()
  })

  it('renders article summary', () => {
    renderWithProviders(
      <ArticleCard
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />
    )

    expect(screen.getByText('Test summary')).toBeInTheDocument()
  })

  it('renders source link', () => {
    renderWithProviders(
      <ArticleCard
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />
    )

    const link = screen.getByRole('link', { name: /Federal Register/i })
    expect(link).toHaveAttribute('href', 'https://example.com')
  })
})
