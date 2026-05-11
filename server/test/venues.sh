#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /venues"
curl -s "$BASE_URL/venues"
echo

echo "GET /venues?search=Accor"
curl -s "$BASE_URL/venues?search=Accor"
echo
