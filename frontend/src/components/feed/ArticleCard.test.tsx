import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ArticleCard } from './ArticleCard'

describe('ArticleCard', () => {
  it('renders article title', () => {
    render(
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
    render(
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
    render(
      <ArticleCard
        title="Test Article"
        summary="Test summary"
        source_url="https://example.com"
        published_at="2024-01-01T00:00:00Z"
      />
    )

    const link = screen.getByRole('link', { name: /Read More/i })
    expect(link).toHaveAttribute('href', 'https://example.com')
  })
})
