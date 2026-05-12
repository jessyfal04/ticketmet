#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /artists"
curl -s "$BASE_URL/artists"
echo

echo "GET /artists?search=NMIX"
curl -s "$BASE_URL/artists?search=NMIX"
echo
