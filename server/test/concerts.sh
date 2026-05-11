#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /concerts"
curl -s "$BASE_URL/concerts"
echo

echo "GET /concerts?artistID=1"
curl -s "$BASE_URL/concerts?artistID=1"
echo

echo "GET /concerts?venueID=1"
curl -s "$BASE_URL/concerts?venueID=1"
echo
