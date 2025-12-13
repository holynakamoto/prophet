# Fix Duplicate Prophet Self-Healing Demo Dashboards

If you see duplicate dashboards in Grafana, here's how to fix it:

## Option 1: Manual Cleanup via Grafana UI

1. Open Grafana: http://localhost:3000
2. Go to **Dashboards** → **Browse**
3. Search for "Prophet Self-Healing Demo"
4. For each duplicate:
   - Click the dashboard
   - Click the gear icon (⚙️) → **Settings**
   - Scroll down and click **Delete Dashboard**
   - Confirm deletion
5. Keep only one dashboard

## Option 2: Using the Cleanup Script

```bash
cd demo/remediation-chaos
./cleanup-dashboards.sh
```

This will:
- Find all dashboards with title "Prophet Self-Healing Demo"
- Keep the first one
- Delete all duplicates

## Option 3: Re-import (Overwrites Existing)

The import script now automatically cleans up duplicates:

```bash
cd demo/remediation-chaos
./import-dashboard.sh
```

This will:
- Clean up any existing duplicates
- Import/update the dashboard (overwrites if exists)

## Prevention

The import script now includes cleanup logic to prevent duplicates. If you see duplicates, it's likely from:
- Multiple manual imports
- Running the import script multiple times before the cleanup was added

The updated import script should prevent this going forward.

