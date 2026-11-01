#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost}"
DIRECT_LEGACY_URL="${DIRECT_LEGACY_URL:-http://localhost:8081}"
LAST_BODY=""

check_status() {
  local expected="$1"
  local label="$2"
  shift 2

  local response actual
  response="$(curl --silent --show-error --write-out $'\n%{http_code}' "$@")"
  actual="${response##*$'\n'}"
  LAST_BODY="${response%$'\n'*}"

  if [[ "$actual" != "$expected" ]]; then
    printf 'FAIL: %s — expected HTTP %s, got HTTP %s\n' "$label" "$expected" "$actual"
    printf '%s\n' "$LAST_BODY"
    exit 1
  fi
  printf 'PASS: %-34s HTTP %s\n' "$label" "$actual"
}

printf 'Generating valid and expired test tokens...\n'
TOKEN=$(MSYS_NO_PATHCONV=1 docker compose run --rm token -subject student -ttl 5m)
EXPIRED_TOKEN=$(MSYS_NO_PATHCONV=1 docker compose run --rm token -subject student -ttl 1s)

if [[ -z "$TOKEN" || -z "$EXPIRED_TOKEN" ]]; then
  printf 'FAIL: token generation returned an empty token\n'
  exit 1
fi
printf 'Waiting for expired token to expire...\n'
sleep 7

check_status 200 'proxy health endpoint' "$BASE_URL/healthz"
check_status 401 'request without token' "$BASE_URL/get"
check_status 401 'malformed token' -H 'Authorization: Bearer invalid-token' "$BASE_URL/get"
check_status 401 'expired token' -H "Authorization: Bearer $EXPIRED_TOKEN" "$BASE_URL/get"
check_status 200 'valid token reaches legacy service' \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/get?source=verification"

if ! grep -q '"args"' <<<"$LAST_BODY" || \
   ! grep -q '"source"' <<<"$LAST_BODY" || \
   ! grep -q '"verification"' <<<"$LAST_BODY"; then
  printf 'FAIL: response body is not the expected go-httpbin /get response\n'
  printf '%s\n' "$LAST_BODY"
  exit 1
fi
printf 'PASS: response body contains expected query data\n'

if curl --fail --silent --connect-timeout 2 --max-time 3 "$DIRECT_LEGACY_URL/get" >/dev/null 2>&1; then
  printf 'FAIL: legacy service is directly accessible at %s\n' "$DIRECT_LEGACY_URL"
  exit 1
fi
printf 'PASS: legacy service is not published on the host\n'
printf '\nAll integration checks passed.\n'
