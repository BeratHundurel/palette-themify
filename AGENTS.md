# Agent Notes

## Quick style pointers

- Svelte uses runes (`$state`/`$effect`/`$derived`).
- Reactive utilities live in `*.svelte.ts`;
- Use `ensureOk` for API error normalization.
- Comment only non-obvious logic.

### Svelte

- Svelte context: https://svelte.dev/llms-medium.txt

### Go

- Utilize 'go doc' and gopls.

### Zig

- Use `zigdoc` to discover current APIs for the Zig standard library and any third-party dependencies before coding.

Examples:

```bash
zigdoc std.fs
zigdoc std.posix.getuid
zigdoc vaxis.Window
```
