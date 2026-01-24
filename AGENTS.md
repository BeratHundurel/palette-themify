# AGENTS.md - Development Guidelines

## Build Commands

### Frontend (SvelteKit)

- `npm run dev` - Development server (port 5173)
- `npm run build` - Production build
- `npm run check` - TypeScript + Svelte checks
- `npm run lint` - ESLint + Prettier checks
- `npm run format` - Format code with Prettier

### Backend (Go)

- `go run .` - Development server (port 8088)
- `go test ./...` - Run all tests
- `make test-one name=TestName` - Run single test
- `make test` - Run all tests with verbose output

### Backend (Zig)

- `zig build run` - Development server
- `zig build test` - Run all tests

## Code Style Guidelines

### Svelte/TypeScript

- Use Svelte 5 runes (`$state`, `$derived`) - avoid legacy stores
- Components: `PascalCase.svelte`
- Reactive utilities: `*.svelte.ts`
- Context files: `context.svelte.ts`
- Colors: `{ hex: string }` format
- Error handling: Use `svelte-french-toast` with loading states

### Formatting

- Tabs, single quotes, no trailing commas
- Print width: 120
- Run `npm run format` and `npm run lint` after changes

### Go

- Standard Go formatting
- Use `make test-one` for single test execution
- Follow existing patterns in handlers and models

### Zig

- Standard Zig formatting (`zig fmt`)
- Use tokamak framework for web server
- Follow existing patterns in API modules

### Critical Rules

- After Svelte/TS changes: ALWAYS run format + lint
- No unnecessary comments - write self-explanatory code
- Use existing API patterns in `lib/api/`
- Follow the UI styling guidelines existing in components if possible or makes sense
