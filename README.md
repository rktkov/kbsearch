# KBSearch — смысловой поиск по методичке I Am Lead

Демо-сервис для практической части мастер-класса 24.10.
Go + SQLite + sqlite-vec, эмбеддинги локально через Ollama (`bge-m3`, 1024).

## Запуск

```bash
# 1. Ollama и модель (один раз)
~/dev/tools/ollama serve &
~/dev/tools/ollama pull bge-m3

# 2. Сборка
go build -o kbsearch ./cmd/kbsearch

# 3. Индекс (≈14 минут на Intel CPU)
./kbsearch ingest --profile fixed --db kb.fixed.sqlite

# 4. Проверки
./kbsearch stats  -db kb.fixed.sqlite
./kbsearch search "fan-out on write" -db kb.fixed.sqlite
go test ./...
./kbsearch eval   -db kb.fixed.sqlite
./kbsearch eval   -db kb.naive.sqlite -profile naive
```

## Профили

| Профиль | Чанкинг | Скор | Дедуп | recall@3 |
|---|---|---|---|---|
| `naive` | фиксированные окна 700 рун | косинус | нет | 0.83 |
| `fixed` | по разделам H1–H3, заголовок в тексте | 0.8·cos + 0.2·лексика | да | 0.92 |

Числа — реальный прогон 23.10: `exhibits/eval-before.txt`, `exhibits/eval-after.txt`.

## Что дальше

- `demo/` — скрипты переключения индекса и прогона S38 (M6)
- `exhibits/` — сырые артефакты агента для показа (M5)
