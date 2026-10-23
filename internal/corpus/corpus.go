// Package corpus загружает документы корпуса с диска.
package corpus

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Doc — один исходный документ. DocID берётся из имени файла без .md,
// потому что именно он показывается в выдаче и используется в eval.
type Doc struct {
	ID      string
	Path    string
	Content string
}

// Load читает все *.md из каталога (не рекурсивно) и сортирует по ID.
func Load(dir string) ([]Doc, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read docs dir %q: %w", dir, err)
	}

	docs := make([]Doc, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}
		docs = append(docs, Doc{
			ID:      strings.TrimSuffix(e.Name(), ".md"),
			Path:    path,
			Content: normalizeNewlines(string(b)),
		})
	}

	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	return docs, nil
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}
