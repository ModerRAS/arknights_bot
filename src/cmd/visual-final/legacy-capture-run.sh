#!/usr/bin/env bash
# Capture the 16 frozen legacy scenes against a locally served fixture set.
#
# Starts the legacy-template fixture service (TestLegacyFixtureService in
# src/core/web), captures every manifest scene with the frozen Playwright client,
# then stops the service. One pass = one fresh server process, so two passes
# never share server state.
#
# Usage:
#   PLAYWRIGHT_MODULE=<npm playwright dir> ./legacy-capture-run.sh <pass-name>
# Env:
#   PLAYWRIGHT_MODULE  required; path to the installed playwright package
#   CAPTURE_OUT_DIR    optional; artifact root (default <repo>/tmp)
#   FIXTURE_ADDR       optional; service address (default 127.0.0.1:38126)
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: PLAYWRIGHT_MODULE=<dir> $0 <pass-name>" >&2
  exit 2
fi
: "${PLAYWRIGHT_MODULE:?set PLAYWRIGHT_MODULE to the installed playwright package dir}"

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
CAPTURE_OUT_DIR="${CAPTURE_OUT_DIR:-$ROOT/tmp}"
FIXTURE_ADDR="${FIXTURE_ADDR:-127.0.0.1:38126}"
CTL="$CAPTURE_OUT_DIR/legacy-capture"
PASS="$1"
OUT="$CAPTURE_OUT_DIR/pass-$PASS"
export NO_PROXY='*' no_proxy='*'

alive() { curl -sf --noproxy '*' -o /dev/null "http://$FIXTURE_ADDR/base"; }

stop_service() {
  touch "$CTL/fixture-service.stop"
  for _ in $(seq 1 60); do
    alive || break
    sleep 1
  done
  rm -f "$CTL/fixture-service.ready" "$CTL/fixture-service.stop" "$CTL/fixture-service.pid"
}

mkdir -p "$CTL"
stop_service
rm -rf "$OUT"
mkdir -p "$OUT"

(cd "$ROOT/src" && LEGACY_FIXTURE_OUT="$CTL" nohup go test ./core/web/ \
  -run TestLegacyFixtureService -count=1 -timeout 60m \
  > "$CTL/fixture-service-$PASS.log" 2>&1 & echo $! > "$CTL/fixture-service.pid")

for _ in $(seq 1 180); do
  [ -f "$CTL/fixture-service.ready" ] && alive && break
  sleep 1
done
alive || { echo "fixture service did not become ready on $FIXTURE_ADDR" >&2; exit 1; }

node "$ROOT/src/cmd/visual-final/legacy-capture.mjs" \
  --out "$OUT" --root "$ROOT" \
  --baseline "$ROOT/src/ggrender/testdata/visual/baseline" \
  --clock-erratum "$ROOT/src/ggrender/testdata/visual/capture-clock-erratum.json" \
  --base-url "http://$FIXTURE_ADDR" --playwright "$PLAYWRIGHT_MODULE" --settle-ms 3000

stop_service