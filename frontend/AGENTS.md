# Frontend Rules

## Project Structure

```
frontend/
├── src/
│   ├── api/                      # API client & utilities
│   ├── components/               # Reusable UI components
│   ├── hook/                     # Feature-level hooks
│   ├── lib/                      # Utility libraries
│   ├── pages/                    # Page components
│   ├── query/                    # TanStack Query code
│   ├── store/                    # Zustand store code
│   ├── styles/                   # Global styles
│   ├── test/                     # Test utilities
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── vite.config.ts
├── tsconfig.json
└── .env.example
```

## Always follow these

- Never use `any` type
- Use non-pluralized naming for all code directories (ex. use hook/ not hooks/)
- Use kebab-case file naming for all .tsx page and components files
- Use camel-case file naming for all other .ts files like hooks, queries, stores, etc

## Implementation Guidelines

- TypeScript throughout
- Zustand for state management
- TanStack Router + Query for routing and data fetching
- shadcn/ui + Tailwind CSS for styling
- Responsive design with loading states and error boundaries

## Patterns

- src/query/: TanStack Query code
- src/store/: Zustand store code
- src/hook/: Feature-level code that orchestrates between queries and stores

## Testing

- Tests required for all features
- Frontend: Vitest + React Testing Library
- Mock external API integrations
- Run tests before commits

## Documentation

- `docs/style.md` - UI component patterns and design decisions

## Commands

### Installation

- `make install-frontend` - Install Node dependencies

### Development

- `make dev-frontend` - Start frontend dev server

### Testing

- `make test-frontend` - Run frontend tests

### Build

- `make build` - Build frontend for production
