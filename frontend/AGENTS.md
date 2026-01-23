# Frontend Rules

## Always follow these:

- Never use `any` type
- Use non-pluralized naming for all code directories (ex. use hook/ not hooks/)
- Use kebab-case file naming for all .tsx page and components files
- Use camel-case file naming for all other .ts files likes hooks, queries, stores, etc

## Patterns

- src/query/: Tanstack query code
- src/store/: Zustand store code
- src/hook/: Feature level code that orchestrates between queries and stores
