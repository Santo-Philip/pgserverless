#!/bin/sh
# Auto-detect free ports for all services
# Usage: ./scripts/ports.sh [env_file]
# If env_file is provided, writes allocated ports to it

set -e

ENV_FILE="${1:-.env}"
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_PATH="$SCRIPT_DIR/$ENV_FILE"

# Default ports
DEFAULT_API=8080
DEFAULT_API_SSL=8443
DEFAULT_GRAFANA=3000
DEFAULT_DASHBOARD=5173

# Read existing value from env file if set
read_env() {
    local key="$1"
    local fallback="$2"
    if [ -f "$ENV_PATH" ]; then
        local val
        val=$(grep "^${key}=" "$ENV_PATH" | head -1 | sed 's/^[^=]*=//')
        echo "${val:-$fallback}"
    else
        echo "$fallback"
    fi
}

# Check if a port is in use (works on Linux)
port_in_use() {
    local port="$1"
    if command -v ss >/dev/null 2>&1; then
        ss -tln "sport = :$port" | grep -q ":$port"
    elif command -v netstat >/dev/null 2>&1; then
        netstat -tln 2>/dev/null | grep -q ":$port "
    else
        # Fallback: try /proc/net/tcp
        awk '{print $2}' /proc/net/tcp 2>/dev/null | grep -qi ":$(printf '%x' "$port")"
    fi
    return $?
}

# Find the next free port starting from a given port
find_free_port() {
    local port="$1"
    while port_in_use "$port"; do
        port=$((port + 1))
    done
    echo "$port"
}

# Set port in env file (create or update)
set_env() {
    local key="$1"
    local val="$2"
    if grep -q "^${key}=" "$ENV_PATH" 2>/dev/null; then
        if [ "$(uname)" = "Darwin" ]; then
            sed -i '' "s|^${key}=.*|${key}=${val}|" "$ENV_PATH"
        else
            sed -i "s|^${key}=.*|${key}=${val}|" "$ENV_PATH"
        fi
    else
        echo "${key}=${val}" >> "$ENV_PATH"
    fi
}

API_PORT=$(read_env "API_PORT" "$DEFAULT_API")
API_PORT_SSL=$(read_env "API_PORT_SSL" "$DEFAULT_API_SSL")
GRAFANA_PORT=$(read_env "GRAFANA_PORT" "$DEFAULT_GRAFANA")
DASHBOARD_PORT=$(read_env "DASHBOARD_PORT" "$DEFAULT_DASHBOARD")

echo "Checking port availability..."

API_PORT=$(find_free_port "$API_PORT")
API_PORT_SSL=$(find_free_port "$API_PORT_SSL")
GRAFANA_PORT=$(find_free_port "$GRAFANA_PORT")
DASHBOARD_PORT=$(find_free_port "$DASHBOARD_PORT")

echo "Allocated ports:"
echo "  API_PORT=$API_PORT"
echo "  API_PORT_SSL=$API_PORT_SSL"
echo "  GRAFANA_PORT=$GRAFANA_PORT"
echo "  DASHBOARD_PORT=$DASHBOARD_PORT"

set_env "API_PORT" "$API_PORT"
set_env "API_PORT_SSL" "$API_PORT_SSL"
set_env "GRAFANA_PORT" "$GRAFANA_PORT"
set_env "DASHBOARD_PORT" "$DASHBOARD_PORT"

echo "Ports written to $ENV_PATH"
