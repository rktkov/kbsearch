#!/usr/bin/env bash
# queries.sh — четыре демо-запроса подряд (запасной вариант, если UI лагает).
# Запросы выбраны так, чтобы первый результат был очевидно правильным.
set -euo pipefail
cd "$(dirname "$0")/.."

queries=(
  "что писать в ADR кроме плюсов"
  "fan-out on write и проблема селебрити"
  "blameless postmortem"
  "сколько стоит пакет теория плюс практика"
)

for q in "${queries[@]}"; do
  ./kbsearch search "$q" --index fixed | head -8
  echo "----------------------------------------"
done
