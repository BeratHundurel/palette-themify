#!/usr/bin/env sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
action=${1:-start}

compose() {
    docker compose \
        --env-file "$repository_root/env.development" \
        -f "$repository_root/docker-compose.yml" \
        -f "$repository_root/docker-compose.observability.yml" \
        "$@"
}

start_stack() {
    compose up --build --detach --wait --wait-timeout 240
    printf '\nThemeSmith is running:\n'
    printf '  App:     http://localhost:5173\n'
    printf '  Jaeger:  http://localhost:16686\n'
    printf '  Go API:  http://localhost:8088/health\n'
    printf '  Zig API: http://localhost:8089/health\n'
}

case "$action" in
    start)
        start_stack
        ;;
    stop)
        compose down
        printf 'ThemeSmith development services stopped. Database data was preserved.\n'
        ;;
    restart)
        compose down
        start_stack
        ;;
    status)
        compose ps
        ;;
    logs)
        compose logs --follow --tail 100
        ;;
    *)
        printf 'Usage: %s {start|stop|restart|status|logs}\n' "$0" >&2
        exit 2
        ;;
esac
