#!/usr/bin/env bash
# pprof_load.sh - concurrent load generator for /api/beers
# Usage:
#   ./pprof_load.sh [concurrency=10] [requests_per_worker=100] [total_requests=0] [url] [duration_seconds=0]
# If total_requests>0 it overrides requests_per_worker and will be divided across workers.
# If duration_seconds>0 the script will run workers for that many seconds (ignores requests_per_worker/total_requests).
set -euo pipefail

CONC=${1:-10}
REQS=${2:-100}
TOTAL=${3:-0}
URL=${4:-'http://localhost:8080/api/beers?offset=0&limit=10'}
DURATION=${5:-30}

# If duration mode is not used, and total_requests provided, distribute across workers
if [[ "$DURATION" -le 0 && "$TOTAL" -gt 0 ]]; then
  # ceil division to distribute total requests across workers
  REQS=$(( (TOTAL + CONC - 1) / CONC ))
fi

if [[ "$DURATION" -gt 0 ]]; then
  echo "Starting load (duration mode): concurrency=$CONC duration=${DURATION}s url=$URL"
else
  echo "Starting load: concurrency=$CONC requests_per_worker=$REQS total=$TOTAL url=$URL"
fi

start_ts=$(date +%s)

worker_duration_mode() {
  local id=$1
  while true; do
    now=$(date +%s)
    elapsed=$((now - start_ts))
    if [[ "$elapsed" -ge "$DURATION" ]]; then
      break
    fi
    curl -s -o /dev/null -w "%{http_code} %{time_total}\n" -X GET "$URL" -H "accept: application/json" || true
  done
}

worker_count_mode() {
  local id=$1
  for i in $(seq 1 "$REQS"); do
    curl -s -o /dev/null -w "%{http_code} %{time_total}\n" -X GET "$URL" -H "accept: application/json" || true
  done
}

# set up signal handler to kill background workers
cleanup() {
  echo "Interrupted, killing workers..."
  pids=$(jobs -p) || true
  if [ -n "${pids:-}" ]; then
    kill $pids 2>/dev/null || true
  fi
  wait
  exit 1
}
trap 'cleanup' SIGINT SIGTERM

for i in $(seq 1 "$CONC"); do
  if [[ "$DURATION" -gt 0 ]]; then
    worker_duration_mode $i &
  else
    worker_count_mode $i &
  fi
done

# wait for background workers
wait

end_ts=$(date +%s)
echo "Done in $((end_ts - start_ts))s"
