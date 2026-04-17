# Agent Notes

## Quick style pointers

- Svelte uses runes (`$state`/`$effect`/`$derived`).
- Reactive utilities live in `*.svelte.ts`; components are PascalCase.
- Use `ensureOk` for API error normalization.
- Comment only non-obvious logic.

## Primary references

- Svelte context: https://svelte.dev/llms-medium.txt
- Frontend patterns: `frontend/src/lib/stores/app.svelte.ts`, `frontend/src/lib/api/base.ts`, `frontend/src/lib/types/color.ts`
- Go API source/docs: `api/go/main.go` and package docs in `api/go/`
- Zig API source/docs: `api/zig/src/palette_api.zig` and related modules in `api/zig/src/`
