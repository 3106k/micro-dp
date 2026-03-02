#!/bin/sh
set -e

# Sync node_modules with package.json on every container start.
# The anonymous volume /app/node_modules persists across rebuilds,
# so new dependencies added to package.json would be missing without this.
npm install --prefer-offline 2>/dev/null || npm install

exec "$@"
