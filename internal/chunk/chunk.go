// Package chunk режет markdown на куски, пригодные для эмбеддинга.
//
// Два режима:
//   - headingAware=true  — по разделам H1–H3, заголовок входит в текст чанка;
//   - headingAware=false — фиксированные окна по рунам (профиль «до»).
package chunk

import (
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
)

// Chunk — единица индексации.
type Chunk struct {
	DocID   string
	Heading string // "C4 → Уровни → Контейнеры"; пусто в naive-режиме
	Body    string // текст без заголовка
	Text    string // то, что уходит в эмбеддинг
	Hash    string // sha1(docID + heading + body)
}

var headingRe = regexp.MustCompile(`^(#{1,3})\s+(.+?)\s*$`)

// Document режет документ на чанки.
func Document(docID, markdown string, maxRunes int, headingAware bool) []Chunk {
	if maxRunes <= 0 {
		maxRunes = 700
	}
	if !headingAware {
		return fixedChunks(docID, markdown, maxRunes)
	}
	return sectionChunks(docID, markdown, maxRunes)
}

// fixedChunks — наивный режим: рубим по рунам, не глядя на структуру.
func fixedChunks(docID, text string, maxRunes int) []Chunk {
	runes := []rune(strings.TrimSpace(text))
	var out []Chunk
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		body := strings.TrimSpace(string(runes[start:end]))
		if body == "" {
			continue
		}
		out = append(out, newChunk(docID, "", body))
	}
	return out
}

// sectionChunks — режим fixed: границы по заголовкам, длинные разделы дробим по абзацам.
func sectionChunks(docID, markdown string, maxRunes int) []Chunk {
	var (
		out     []Chunk
		path    [3]string
		body    []string
		hasBody bool
	)

	flush := func() {
		if !hasBody {
			return
		}
		heading := joinPath(path)
		for _, piece := range splitBody(strings.Join(body, "\n"), maxRunes) {
			out = append(out, newChunk(docID, heading, piece))
		}
		body, hasBody = nil, false
	}

	for _, line := range strings.Split(markdown, "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			flush()
			level := len(m[1])
			path[level-1] = strings.TrimSpace(m[2])
			for i := level; i < len(path); i++ {
				path[i] = ""
			}
			continue
		}
		body = append(body, line)
		hasBody = true
	}
	flush()

	return out
}

// splitBody делит текст раздела на куски не длиннее maxRunes, стараясь
// резать по границам абзацев.
func splitBody(text string, maxRunes int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len([]rune(text)) <= maxRunes {
		return []string{text}
	}

	var (
		pieces  []string
		current strings.Builder
		curLen  int
	)

	push := func() {
		s := strings.TrimSpace(current.String())
		if s != "" {
			pieces = append(pieces, s)
		}
		current.Reset()
		curLen = 0
	}

	for _, para := range strings.Split(text, "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		paraRunes := []rune(para)

		// Абзац сам по себе больше лимита — рубим жёстко.
		if len(paraRunes) > maxRunes {
			push()
			for start := 0; start < len(paraRunes); start += maxRunes {
				end := start + maxRunes
				if end > len(paraRunes) {
					end = len(paraRunes)
				}
				pieces = append(pieces, strings.TrimSpace(string(paraRunes[start:end])))
			}
			continue
		}

		if curLen > 0 && curLen+2+len(paraRunes) > maxRunes {
			push()
		}
		if curLen > 0 {
			current.WriteString("\n\n")
			curLen += 2
		}
		current.WriteString(para)
		curLen += len(paraRunes)
	}
	push()

	return pieces
}

func newChunk(docID, heading, body string) Chunk {
	text := body
	if heading != "" {
		text = heading + "\n\n" + body
	}
	sum := sha1.Sum([]byte(docID + "\x00" + heading + "\x00" + body))
	return Chunk{
		DocID:   docID,
		Heading: heading,
		Body:    body,
		Text:    text,
		Hash:    hex.EncodeToString(sum[:]),
	}
}

func joinPath(path [3]string) string {
	parts := make([]string, 0, len(path))
	for _, p := range path {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " → ")
}
