#!/bin/bash
# Persistent Headlamp port-forward script
# This will automatically restart the port-forward if it dies

echo "Starting Headlamp port-forward on http://localhost:4466"
echo "Press Ctrl+C to stop"
echo ""

while true; do
  echo "$(date): Starting port-forward..."
  kubectl port-forward -n headlamp svc/headlamp 4466:80
  EXIT_CODE=$?
  
  if [ $EXIT_CODE -ne 0 ]; then
    echo "$(date): Port-forward exited with code $EXIT_CODE, restarting in 2 seconds..."
    sleep 2
  else
    echo "$(date): Port-forward ended normally, restarting..."
    sleep 1
  fi
done

