#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /api/artists"
curl -s "$BASE_URL/api/artists"
echo

echo "GET /api/artists?search=NMIX"
curl -s "$BASE_URL/api/artists?search=NMIX"
echo
