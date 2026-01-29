#!/bin/bash
set -e

cd "$(dirname "$0")/.."

mkdir -p bin

go build -ldflags "-s -w" -o bin/jobs ./cmd/jobs
