#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Load environment variables
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

./bin/scraper --once
