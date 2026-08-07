# ThemeSmith

ThemeSmith turns images into reusable color palettes and editor themes. It is available as a web application and a Wails desktop application, with theme generation handled by a dedicated Zig service and persistence, accounts, and sharing handled by Go.

## Features

- Extract a color palette from an uploaded image.
- Apply, reorder, save, and copy generated colors.
- Generate and customize VS Code-family and Zed themes.
- Install supported themes directly from the desktop application.
- Save palettes and themes to an account and synchronize preferences.
- Share palettes and themes through the searchable community gallery.
- Inspect application and browser errors through OpenTelemetry and Jaeger.

## Architecture

| Component  | Technology                      | Responsibility                                                       |
| ---------- | ------------------------------- | -------------------------------------------------------------------- |
| `frontend` | SvelteKit, Svelte 5, TypeScript | Web and desktop UI                                                   |
| `api/go`   | Go, Gin, GORM                   | Accounts, persistence, sharing, preferences, and telemetry ingestion |
| `api/zig`  | Zig, http.zig, zigimg           | Palette extraction and editor-theme generation                       |
| `desktop`  | Wails 3, Go                     | Native desktop shell and direct editor installation                  |
| `postgres` | PostgreSQL                      | User, palette, theme, and preference storage                         |
| `jaeger`   | Jaeger                          | Local OpenTelemetry collection and trace/error viewer                |

## Quick start

Docker Desktop with Docker Compose is the only requirement for the complete development stack.

On Windows PowerShell:

```powershell
.\scripts\dev.ps1 start
```

On macOS or Linux:

```sh
sh ./scripts/dev.sh start
```

The initial start downloads dependencies and may take a few minutes. Subsequent starts reuse Docker and language caches.

| Service        | URL                          |
| -------------- | ---------------------------- |
| ThemeSmith     | http://localhost:5173        |
| Jaeger         | http://localhost:16686       |
| Go API health  | http://localhost:8088/health |
| Zig API health | http://localhost:8089/health |
| PostgreSQL     | `localhost:5433`             |

## Managing the development stack

The helper accepts `start`, `stop`, `restart`, `status`, and `logs`:

```powershell
.\scripts\dev.ps1 status
.\scripts\dev.ps1 logs
.\scripts\dev.ps1 restart
.\scripts\dev.ps1 stop
```

Use `sh ./scripts/dev.sh <action>` for the equivalent Unix command. `stop` removes the development containers and network but preserves the PostgreSQL volume and its data.

To run without Jaeger and browser error reporting:

```sh
docker compose --env-file env.development up --build -d
```

## Configuration

Safe local defaults are tracked in `env.development`. Change its published ports and browser-facing URLs when they conflict with other services on your machine.

Desktop builds use the `THEMESMITH_API_BASE_URL` and `THEMESMITH_ZIG_API_BASE_URL` Task variables. Set `THEMESMITH_TELEMETRY_ENABLED=true` at build time to enable browser-side error reporting in the desktop application.

## Error telemetry

The development helper starts Jaeger and enables OpenTelemetry automatically. The application records:

- Go request panics and HTTP 5xx responses;
- browser uncaught errors and unhandled promise rejections;
- failed browser API responses and `console.error` calls;
- a trace ID in the Go API's `X-Trace-ID` response header.

Open http://localhost:16686, select the `themesmith-api` service, and filter for error traces. The local Jaeger instance uses in-memory storage, so traces are discarded when its container is removed.

For hosted telemetry, set `OTEL_EXPORTER_OTLP_ENDPOINT` and any provider-required standard OpenTelemetry variables such as `OTEL_EXPORTER_OTLP_HEADERS`. Browser telemetry is controlled at build time with `VITE_TELEMETRY_ENABLED`.

## Development checks

Frontend:

```sh
cd frontend
npm ci
npm run check
npm run lint
npm run test
```

Go API:

```sh
cd api/go
go test ./...
go vet ./...
```

Zig API:

```sh
cd api/zig
zig build test
```