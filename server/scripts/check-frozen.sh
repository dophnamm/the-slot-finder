#!/bin/sh
# Frozen-file rules (review level L0, no AI): files that are generated or already applied must
# match their source. Compares the working tree (commits plus uncommitted work) with the base.
# Prints PASS, FAIL or SKIP per rule; exits 1 when any rule fails. A SKIP never fails.
# Usage: make check-frozen BASE=main   (the Makefile passes: base, migrations dir, swag command)

BASE="$1"
MIGRATIONS_DIR="$2"
SWAG="$3"
cd "$(dirname "$0")/.." || exit 2

MERGE_BASE="$(git merge-base "$BASE" HEAD)" || { echo "FAIL no merge base with $BASE"; exit 1; }
STATUS=0

# Applied migrations are append-only: a file that exists on the base must not change or go away.
# atlas.sum changes with every new migration, so it is left out.
CHANGED="$(git diff --name-status "$MERGE_BASE" -- "$MIGRATIONS_DIR" | grep -E '^[MDR]' | grep -v 'atlas\.sum$')"
if [ -n "$CHANGED" ]; then
  echo "FAIL migrations: applied migration files changed; add a new migration instead"
  echo "$CHANGED"
  STATUS=1
else
  echo "PASS migrations: applied migration files unchanged"
fi

# Swagger in docs/ is generated from the handler annotations: regenerate it into a temp dir and
# compare, so a hand edit or a forgotten `make swagger-load` shows up.
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
# SWAG is a multi-word command (go run ...), so it is left unquoted on purpose.
if ! $SWAG --version >/dev/null 2>&1; then
  echo "SKIP swagger: cannot run swag (needs Go, and network on the first run)"
elif ! make -s swagger-load SWAGGER_OUT="$TMP/docs" >"$TMP/swag.log" 2>&1; then
  echo "FAIL swagger: swag could not parse the handler annotations"
  tail -20 "$TMP/swag.log"
  STATUS=1
elif diff -r "$TMP/docs" docs >"$TMP/diff.log"; then
  echo "PASS swagger: docs/ matches the handler annotations"
else
  echo "FAIL swagger: docs/ differs from the annotations; run make swagger-load (never edit docs/ by hand)"
  head -40 "$TMP/diff.log"
  STATUS=1
fi

exit $STATUS
