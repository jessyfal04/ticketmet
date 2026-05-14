#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /api/concerts"
curl -s "$BASE_URL/api/concerts"
echo

echo "GET /api/concerts?artistID=1"
curl -s "$BASE_URL/api/concerts?artistID=1"
echo

echo "GET /api/concerts?venueID=1"
curl -s "$BASE_URL/api/concerts?venueID=1"
echo
