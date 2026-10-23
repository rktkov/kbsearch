#!/usr/bin/env bash
# reset.sh empty|naive|fixed — переключает рабочий индекс (kb.sqlite).
set -euo pipefail
cd "$(dirname "$0")/.."

mode="${1:-empty}"
rm -f kb.sqlite kb.sqlite-shm kb.sqlite-wal

case "$mode" in
  empty)
    echo "index: empty (kb.sqlite отсутствует — stats покажет 30 документов и 0 чанков)"
    ;;
  naive|fixed)
    if [ ! -f "kb.$mode.sqlite" ]; then
      echo "нет файла kb.$mode.sqlite" >&2
      exit 1
    fi
    cp "kb.$mode.sqlite" kb.sqlite
    echo "index: $mode"
    ;;
  *)
    echo "usage: reset.sh empty|naive|fixed" >&2
    exit 2
    ;;
esac
