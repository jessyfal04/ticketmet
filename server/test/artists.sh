#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "GET /artists"
curl -s "$BASE_URL/artistes"
echo

echo "GET /artists?search=NMIX"
curl -s "$BASE_URL/artistes?search=NMIX"
echo
