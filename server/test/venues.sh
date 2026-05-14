#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /api/venues"
curl -s "$BASE_URL/api/venues"
echo

echo "GET /api/venues?search=Accor"
curl -s "$BASE_URL/api/venues?search=Accor"
echo
