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

## UI Component Styling Guidelines

### Design System Principles

#### Color Palette

- **Primary Brand**: `text-brand` class for primary actions, headers, and accents
- **Secondary**: Zinc variants (`zinc-100` to `zinc-900`) for backgrounds, borders, and text
- **Interactive States**:
  - Hover: `hover:bg-zinc-800/50` for backgrounds, `hover:border-brand/50` for borders
  - Focus: `focus:border-brand/50` for inputs
  - Active: `active:scale-95` for buttons

#### Typography Hierarchy

- **Headers**: `text-sm font-semibold tracking-wide uppercase text-brand` (section headers)
- **Main Title**: `text-2xl font-semibold text-brand` (modal titles)
- **Button Labels**: `text-sm font-medium` (primary), `text-sm font-semibold` (secondary)
- **Small Text**: `text-xs font-medium text-zinc-500` (labels, metadata)
- **Code/IDs**: `font-mono` with appropriate sizing

#### Spacing System

- **Container Padding**: `px-6 py-5` (header/footer), `px-6 py-6` (content)
- **Section Spacing**: `mb-8` (major sections), `mb-6` (subsections), `mb-4` (minor)
- **Element Gaps**: `gap-4` (button groups), `gap-3` (icon-text pairs)
- **Internal Padding**: `p-4` (cards), `p-3` (compact sections)

#### Component Patterns

##### Buttons

```svelte
<!-- Primary Action Button -->
<button class="bg-brand shadow-brand/20 hover:shadow-brand/40 rounded-lg px-5 py-2.5 text-sm font-semibold text-zinc-900 transition-all duration-300 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:scale-100">
  Primary Action
</button>

<!-- Secondary Action Button -->
<button class="rounded-lg border border-zinc-600 px-5 py-2.5 text-sm font-medium text-zinc-300 transition-all duration-300 hover:border-brand/50 hover:bg-zinc-800/50">
  Secondary Action
</button>

<!-- Icon/Toolbar Button -->
<button class="toolbar-button-base">
  <!-- Content -->
</button>

<!-- Subtle Action Button -->
<button class="hover:text-brand rounded-lg p-2 text-zinc-400 transition-all duration-300 hover:bg-zinc-800/50">
  <!-- Icon -->
</button>
```

##### Input Fields

```svelte
<input
  class="focus:border-brand/50 w-full rounded-lg border border-zinc-700 bg-zinc-900 py-2 pr-10 pl-10 text-sm text-zinc-300 placeholder-zinc-400 focus:outline-none transition-all duration-300"
  placeholder="Placeholder text"
/>
```

##### Modal/Dialog Container

```svelte
<div class="animate-scale-in relative w-full max-w-4xl rounded-xl border border-brand/50 bg-zinc-900 shadow-brand/20 shadow-2xl">
  <!-- Content -->
</div>
```

##### Card Components

```svelte
<div class="hover:border-brand/50 hover:shadow-brand/10 overflow-hidden rounded-xl border-2 border-zinc-700/50 bg-zinc-800/30 backdrop-blur-sm transition-all duration-300">
  <!-- Card content -->
</div>
```

#### Animation & Transitions

- **Standard**: `transition-all duration-300` for all interactive elements
- **Hover Effects**: `hover:scale-105` (buttons), `hover:opacity-80` (images), `hover:bg-zinc-800/50` (backgrounds)
- **Modal Entry**: `animate-scale-in` class
- **Loading Spinners**: `text-brand border-t-brand animate-spin`

#### Icon Standards

- **Size**: `h-4 w-4` (small), `h-5 w-5` (medium), `height="18px" width="18px"` (toolbar)
- **Color**: `text-zinc-400` (default), `text-brand` (active/selected), `hover:text-brand` (hover)
- **Structure**: Consistent `fill="currentColor"` or `stroke="currentColor"`

#### Border & Shadow System

- **Borders**: `border-zinc-700/50` (default), `border-brand/50` (selected/focus)
- **Shadows**: `shadow-brand/20` (subtle), `shadow-brand/40` (hover), `shadow-2xl` (modals)
- **Backgrounds**: `bg-zinc-900` (primary), `bg-zinc-800/50` (subtle), `bg-zinc-800/30` (cards)

#### Component-Specific Guidelines

##### Search Components

- Use consistent `max-w-md` for search forms
- Add `group` class to form for coordinated hover states
- Search button should use primary button styling
- Clear button should use subtle icon button pattern

##### Toolbar Components

- All toolbar buttons use `toolbar-button-base` class
- No custom classes that override the base styling
- Consistent icon sizing and positioning
- No emoji icons - use SVG icons only

##### Modal Components

- Use consistent modal container styling
- Header/footer with `px-6 py-5`
- Content area with `px-6 py-6`
- Consistent section spacing (`mb-6`)

#### Consistency Rules

1. **Always use `transition-all duration-300`** for hover states
2. **Consistent hover backgrounds**: `hover:bg-zinc-800/50`
3. **Consistent hover borders**: `hover:border-brand/50`
4. **Consistent brand color application**: Headers, primary actions, selected states
5. **Standardized padding patterns**: No arbitrary values
6. **Consistent font hierarchy**: Don't mix `font-medium` and `font-semibold` randomly

### Critical Rules

- After Svelte/TS changes: ALWAYS run format + lint
- No unnecessary comments - write self-explanatory code
- Use existing API patterns in `lib/api/`
- Canvas coordinates require scaling consideration
- Follow the UI styling guidelines above for all component changes
- Always use consistent hover states and transitions
- Maintain the established color palette and spacing system
