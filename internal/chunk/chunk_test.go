package chunk

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDocumentSplitsLargeText(t *testing.T) {
	doc := strings.Repeat("Это предложение про архитектуру. ", 200) // ~6600 рун
	chunks := Document("doc", doc, 700, false)
	if len(chunks) < 6 {
		t.Fatalf("ожидали ≥6 чанков, получили %d", len(chunks))
	}
}

func TestChunkLengthIsCountedInRunes(t *testing.T) {
	// 2000 кириллических символов = 4000 байт. Если считать байты, чанков будет вдвое больше.
	doc := strings.Repeat("я", 2000)
	chunks := Document("doc", doc, 700, false)
	if len(chunks) != 3 {
		t.Fatalf("ожидали 3 чанка по 700 рун, получили %d", len(chunks))
	}
	for _, c := range chunks {
		if n := utf8.RuneCountInString(c.Body); n > 750 {
			t.Fatalf("чанк длиннее лимита: %d рун", n)
		}
	}
}

func TestHeadingAwareKeepsStructure(t *testing.T) {
	md := `# C4

Введение в модель.

## Уровни

Первый уровень — контекст.

## Контейнеры

Второй уровень — контейнеры.
`
	chunks := Document("c4", md, 700, true)
	if len(chunks) < 3 {
		t.Fatalf("ожидали ≥3 чанка по разделам, получили %d", len(chunks))
	}

	var found bool
	for _, c := range chunks {
		if c.Heading == "C4 → Контейнеры" {
			found = true
			if !strings.HasPrefix(c.Text, "C4 → Контейнеры\n\n") {
				t.Fatalf("заголовок не попал в текст чанка: %q", c.Text)
			}
			if !strings.Contains(c.Body, "Второй уровень") {
				t.Fatalf("тело чанка не то: %q", c.Body)
			}
		}
	}
	if !found {
		t.Fatal("не нашли чанк с heading path «C4 → Контейнеры»")
	}
}

func TestChunksAreDeterministic(t *testing.T) {
	md := "# Раздел\n\nТекст про инварианты и агрегаты.\n"
	a := Document("doc", md, 700, true)
	b := Document("doc", md, 700, true)
	if len(a) != len(b) {
		t.Fatalf("разное число чанков: %d и %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Hash != b[i].Hash {
			t.Fatalf("хеши расходятся на чанке %d", i)
		}
	}
}

func TestNaiveModeIgnoresHeadings(t *testing.T) {
	md := "# Заголовок\n\nТекст.\n"
	naive := Document("doc", md, 700, false)
	for _, c := range naive {
		if c.Heading != "" {
			t.Fatalf("в naive-режиме heading должен быть пустым, получили %q", c.Heading)
		}
	}
}
