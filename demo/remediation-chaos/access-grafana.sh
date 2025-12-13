#!/usr/bin/env bash
# Quick script to access Grafana dashboard
# Usage: GRAFANA_USER=admin GRAFANA_PASS=... ./access-grafana.sh

echo "📊 Accessing Prophet Self-Healing Demo Dashboard"
echo ""

GRAFANA_USER=${GRAFANA_USER:-admin}
GRAFANA_PASS=${GRAFANA_PASS:-admin}

# Check if port-forward is already running
PF_PID=$(ps aux | grep "kubectl port-forward.*grafana" | grep -v grep | awk '{print $2}' | head -1)

if [ -z "$PF_PID" ]; then
    echo "🔗 Starting port-forward to Grafana..."
    kubectl port-forward -n monitoring svc/grafana 3000:3000 > /tmp/grafana-pf.log 2>&1 &
    PF_PID=$!
    echo "   Port-forward started (PID: $PF_PID)"
    sleep 3
else
    echo "✅ Port-forward already running (PID: $PF_PID)"
fi

# Check if Grafana is accessible
if curl -s -f http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana is accessible at http://localhost:3000"
else
    echo "⚠️  Grafana not responding. Waiting a bit longer..."
    sleep 5
    if curl -s -f http://localhost:3000/api/health > /dev/null; then
        echo "✅ Grafana is now accessible"
    else
        echo "❌ Grafana is not accessible. Check:"
        echo "   kubectl get pods -n monitoring -l app=grafana"
        echo "   kubectl logs -n monitoring -l app=grafana"
        exit 1
    fi
fi

# Get dashboard UID
DASHBOARD_UID=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASS" http://localhost:3000/api/search?query=prophet-self-healing | jq -r '.[] | select(.title=="Prophet Self-Healing Demo") | .uid' 2>/dev/null | head -1)

if [ -n "$DASHBOARD_UID" ] && [ "$DASHBOARD_UID" != "null" ]; then
    DASHBOARD_URL="http://localhost:3000/d/$DASHBOARD_UID/prophet-self-healing-demo"
    echo ""
    echo "✅ Dashboard found!"
    echo ""
    echo "🔗 Dashboard URL: $DASHBOARD_URL"
    echo ""
    echo "📋 Login credentials:"
    echo "   Username: $GRAFANA_USER"
    echo "   Password: (use GRAFANA_PASS env var)"
    echo ""
    echo "💡 Opening in browser..."
    
    # Try to open in browser (works on macOS and Linux)
    if command -v open > /dev/null; then
        open "$DASHBOARD_URL"
    elif command -v xdg-open > /dev/null; then
        xdg-open "$DASHBOARD_URL"
    else
        echo "   Copy and paste the URL above into your browser"
    fi
else
    echo ""
    echo "⚠️  Dashboard not found. Importing..."
    ./import-dashboard.sh
    echo ""
    echo "   Then run this script again or visit: http://localhost:3000"
fi

echo ""
echo "🛑 To stop port-forward: kill $PF_PID"

