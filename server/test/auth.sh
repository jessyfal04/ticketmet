#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="$(mktemp)"
EMAIL="auth-test-$(date +%s)-$$@example.com"
PASSWORD="password123"

trap 'rm -f "$COOKIE_JAR"' EXIT

echo "POST /api/auth/register"
register_body="$(curl -s -w '\n%{http_code}' -c "$COOKIE_JAR" \
	-H 'Content-Type: application/json' \
	-d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
	"$BASE_URL/api/auth/register")"
test "${register_body##*$'\n'}" = "201"
echo "${register_body%$'\n'*}"

echo "GET /api/auth/me"
me_body="$(curl -s -w '\n%{http_code}' -b "$COOKIE_JAR" "$BASE_URL/api/auth/me")"
test "${me_body##*$'\n'}" = "200"
echo "${me_body%$'\n'*}"

echo "POST /api/auth/logout"
logout_code="$(curl -s -o /dev/null -w '%{http_code}' -b "$COOKIE_JAR" -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/logout")"
test "$logout_code" = "204"

echo "POST /api/auth/login"
login_body="$(curl -s -w '\n%{http_code}' -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
	-H 'Content-Type: application/json' \
	-d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
	"$BASE_URL/api/auth/login")"
test "${login_body##*$'\n'}" = "200"
echo "${login_body%$'\n'*}"

echo "GET /api/auth/email-exists"
exists_body="$(curl -s "$BASE_URL/api/auth/email-exists?email=$EMAIL")"
echo "$exists_body" | grep -q '"exists":true'
echo "$exists_body"

echo "DELETE /api/auth/unregister"
unregister_code="$(curl -s -o /dev/null -w '%{http_code}' -b "$COOKIE_JAR" \
	-X DELETE -H 'Content-Type: application/json' \
	-d "{\"password\":\"$PASSWORD\"}" \
	"$BASE_URL/api/auth/unregister")"
test "$unregister_code" = "204"
