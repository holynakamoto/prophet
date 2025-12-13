#!/usr/bin/env bash
# Clean up duplicate Prophet Self-Healing Demo dashboards
# Usage: ./cleanup-dashboards.sh [grafana-url] [admin-password]

GRAFANA_URL=${1:-http://localhost:3000}
GRAFANA_PASS=${2:-admin}

echo "🧹 Cleaning up duplicate Prophet Self-Healing Demo dashboards..."

# Check if Grafana is accessible
if ! curl -s -f -u admin:$GRAFANA_PASS "$GRAFANA_URL/api/health" > /dev/null; then
    echo "⚠️  Cannot connect to Grafana at $GRAFANA_URL"
    exit 1
fi

# Find all dashboards with "Prophet Self-Healing Demo" in the title
DASHBOARDS=$(curl -s -u admin:$GRAFANA_PASS "$GRAFANA_URL/api/search?type=dash-db&query=prophet" | jq -r 'if type == "array" then .[] | select(.title == "Prophet Self-Healing Demo") | .uid else empty end' 2>/dev/null)

if [ -z "$DASHBOARDS" ]; then
    echo "✅ No duplicate dashboards found"
    exit 0
fi

# Count duplicates
COUNT=$(echo "$DASHBOARDS" | wc -l | tr -d ' ')
echo "   Found $COUNT dashboard(s) with title 'Prophet Self-Healing Demo'"

# Keep the first one, delete the rest
FIRST=true
KEPT_UID=""
DELETED_COUNT=0

echo "$DASHBOARDS" | while read uid; do
    if [ -n "$uid" ] && [ "$uid" != "null" ]; then
        if [ "$FIRST" = true ]; then
            echo "   ✓ Keeping dashboard: $uid"
            KEPT_UID="$uid"
            FIRST=false
        else
            echo "   🗑️  Deleting duplicate dashboard: $uid"
            curl -s -X DELETE -u admin:$GRAFANA_PASS "$GRAFANA_URL/api/dashboards/uid/$uid" > /dev/null 2>&1
            if [ $? -eq 0 ]; then
                DELETED_COUNT=$((DELETED_COUNT + 1))
                echo "      ✓ Deleted"
            else
                echo "      ⚠️  Failed to delete"
            fi
        fi
    fi
done

echo ""
if [ "$DELETED_COUNT" -gt 0 ]; then
    echo "✅ Cleanup complete! Deleted $DELETED_COUNT duplicate(s)"
    if [ -n "$KEPT_UID" ]; then
        echo "   Remaining dashboard: $GRAFANA_URL/d/$KEPT_UID/prophet-self-healing-demo"
    fi
else
    echo "✅ No duplicates to delete"
fi

