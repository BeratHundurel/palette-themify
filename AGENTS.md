# Project Guidelines

## Code Style

- Svelte uses runes ($state/$effect/$derived) in stores; follow patterns in [frontend/src/lib/stores/app.svelte.ts](frontend/src/lib/stores/app.svelte.ts).
- API helpers rely on `ensureOk` for error normalization; see [frontend/src/lib/api/base.ts](frontend/src/lib/api/base.ts).
- Components are PascalCase and reactive utilities live in \*.svelte.ts (see [frontend/src/lib/stores](frontend/src/lib/stores)).
- Colors are shaped as { hex: string } (see [frontend/src/lib/types/color.ts](frontend/src/lib/types/color.ts)).
- Formatting: tabs, single quotes, no trailing commas, print width 120; after Svelte/TS changes run `npm run format` and `npm run lint`.

## Architecture

- Frontend is SvelteKit under frontend/src; HTTP is isolated in [frontend/src/lib/api](frontend/src/lib/api).
- Go API (port 8088) owns auth, palettes, workspaces, apply-palette, and Wallhaven proxy routes in [api/go/main.go](api/go/main.go).
- Zig API (port 8089) handles palette extraction and theme generation via Tokamak in [api/zig/src/main.zig](api/zig/src/main.zig) with core logic in [api/zig/src/palette_api.zig](api/zig/src/palette_api.zig).

## Build and Test

- Frontend: `npm run dev`, `npm run build`, `npm run check`, `npm run lint`, `npm run format` (run in frontend/).
- Go API: `go run .`, `go test ./...`, `make test-one name=TestName`, `make test` (run in api/go/).
- Zig API: `zig build run`, `zig build test` (run in api/zig/).

## Project Conventions

- API base URLs are `VITE_API_BASE_URL` and `VITE_ZIG_API_BASE_URL` with localhost defaults in [frontend/src/lib/api/base.ts](frontend/src/lib/api/base.ts).
- Auth headers attach `Authorization: Bearer <token>` via [frontend/src/lib/api/auth.ts](frontend/src/lib/api/auth.ts).
- Wallhaven proxy forwards `X-API-Key` through the Go API in [api/go/wallhaven_handlers.go](api/go/wallhaven_handlers.go).
- Workspace sharing uses /workspaces/:id/share and /shared token flow (see [api/go/workspace_handlers.go](api/go/workspace_handlers.go) and [frontend/src/lib/api/workspace.ts](frontend/src/lib/api/workspace.ts)).

## Integration Points

- Frontend calls Zig endpoints /extract-palette, /generate-theme, /generate-overridable in [frontend/src/lib/api/palette.ts](frontend/src/lib/api/palette.ts) and [frontend/src/lib/api/theme.ts](frontend/src/lib/api/theme.ts).
- Frontend calls Go endpoints for palettes/workspaces/wallhaven/apply-palette in [frontend/src/lib/api/palette.ts](frontend/src/lib/api/palette.ts), [frontend/src/lib/api/workspace.ts](frontend/src/lib/api/workspace.ts), [frontend/src/lib/api/wallhaven.ts](frontend/src/lib/api/wallhaven.ts), and [frontend/src/lib/api/theme.ts](frontend/src/lib/api/theme.ts).
- Go API proxies Wallhaven requests in [api/go/wallhaven_handlers.go](api/go/wallhaven_handlers.go).

## Security

- JWT auth middleware validates `Authorization: Bearer` in [api/go/auth.go](api/go/auth.go); the frontend stores the token in localStorage via [frontend/src/lib/api/auth.ts](frontend/src/lib/api/auth.ts).
- Demo login/user bootstrap and sample palettes are created in [api/go/auth.go](api/go/auth.go).
- Database config is loaded from env in [api/go/db.go](api/go/db.go); it currently prints the DB password during init.
