# Project Guidelines

## Code Style

- Svelte uses runes (`$state`/`$effect`/`$derived`) — follow patterns in [app.svelte.ts](frontend/src/lib/stores/app.svelte.ts).
- Reactive utilities live in `*.svelte.ts`; components are PascalCase.
- Colors are `{ hex: string }` — see [color.ts](frontend/src/lib/types/color.ts).
- Use `ensureOk` for API error normalization — see [base.ts](frontend/src/lib/api/base.ts).
- Comment only non-obvious logic. Prefer simple, direct solutions.

## Architecture

- **Frontend** (SvelteKit, `frontend/src`): all HTTP calls isolated in `frontend/src/lib/api/`.
- **Go API** (port 8088): auth, palettes, workspaces, apply-palette, Wallhaven proxy — [main.go](api/go/main.go).
- **Zig API** (port 8089): palette extraction and theme generation via Tokamak — [palette_api.zig](api/zig/src/palette_api.zig).
