#!/usr/bin/env bash
# Import Prophet Self-Healing Demo Dashboard into Grafana
# Usage: GRAFANA_USER=admin GRAFANA_PASS=... ./import-dashboard.sh [grafana-url]

GRAFANA_URL=${1:-http://localhost:3000}
GRAFANA_USER=${GRAFANA_USER:-admin}
GRAFANA_PASS=${GRAFANA_PASS:-${2:-admin}}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DASHBOARD_FILE="$SCRIPT_DIR/../../monitoring/grafana/dashboards/prophet-self-healing-demo.json"

echo "📊 Importing Prophet Self-Healing Demo Dashboard..."
echo "   Grafana URL: $GRAFANA_URL"

# Check if Grafana is reachable (do not require auth for this check)
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$GRAFANA_URL/api/health" || true)
if [ "$HEALTH_STATUS" != "200" ]; then
    echo "⚠️  Grafana is not reachable at $GRAFANA_URL (HTTP $HEALTH_STATUS)"
    echo "   Make sure Grafana is running and port-forwarded:"
    echo "   kubectl port-forward -n monitoring svc/grafana 3000:3000"
    exit 1
fi

# Validate credentials early for a clearer error than "cannot connect"
AUTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/user" || true)
if [ "$AUTH_STATUS" != "200" ]; then
    echo "❌ Grafana authentication failed (HTTP $AUTH_STATUS)"
    echo "   Set credentials via env vars:"
    echo "   GRAFANA_USER=admin GRAFANA_PASS='...' ./import-dashboard.sh"
    exit 1
fi

# Read dashboard JSON
if [ ! -f "$DASHBOARD_FILE" ]; then
    echo "❌ Dashboard file not found: $DASHBOARD_FILE"
    exit 1
fi

# Extract dashboard from nested structure if needed
if cat "$DASHBOARD_FILE" | jq -e '.dashboard' > /dev/null 2>&1; then
    DASHBOARD_JSON=$(cat "$DASHBOARD_FILE" | jq -c '.dashboard')
else
    DASHBOARD_JSON=$(cat "$DASHBOARD_FILE" | jq -c '.')
fi

# Clean up any existing duplicates first
if [ -f "$SCRIPT_DIR/cleanup-dashboards.sh" ]; then
    "$SCRIPT_DIR/cleanup-dashboards.sh" "$GRAFANA_URL" "$GRAFANA_PASS" > /dev/null 2>&1
fi

# Import dashboard via API (overwrite will update the existing one)
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -u "$GRAFANA_USER:$GRAFANA_PASS" \
    -d "{\"dashboard\":$DASHBOARD_JSON,\"overwrite\":true}" \
    "$GRAFANA_URL/api/dashboards/db")

if echo "$RESPONSE" | jq -e '.uid' > /dev/null 2>&1; then
    DASHBOARD_UID=$(echo "$RESPONSE" | jq -r '.uid')
    echo "✅ Dashboard imported successfully!"
    echo "   View at: $GRAFANA_URL/d/$DASHBOARD_UID/prophet-self-healing-demo"
else
    echo "❌ Failed to import dashboard:"
    echo "$RESPONSE" | jq -r '.message // .'
    exit 1
fi

