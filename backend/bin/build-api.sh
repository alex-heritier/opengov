#!/bin/bash
set -e

cd "$(dirname "$0")/.."

mkdir -p dist

go build -ldflags "-s -w" -o dist/api ./cmd/api
