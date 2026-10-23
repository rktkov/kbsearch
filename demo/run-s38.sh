#!/usr/bin/env bash
# run-s38.sh — вся последовательность кульминации (S38) одной командой.
# Используется на репетиции и как бэкап, если в эфире что-то пойдёт не так.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "=== 1. ТЕСТЫ ==="
go test ./... 2>&1 | grep -E "^(ok|FAIL|---)" || true

echo
echo "=== 2. ДО: наивный индекс ==="
./kbsearch eval --index naive | tail -16

echo
echo "=== 3. ПОСЛЕ: чанки по разделам + гибридный скор ==="
./kbsearch eval --index fixed | tail -16

echo
echo "=== 4. ЖИВОЙ ПОИСК ==="
./kbsearch search "что писать в ADR кроме плюсов" --index fixed | head -10
