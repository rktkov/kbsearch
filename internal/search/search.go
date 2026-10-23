// Package search — поиск: KNN + гибридный скор + дедупликация по документу.
package search

import (
	"context"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rktkov/kbsearch/internal/config"
	"github.com/rktkov/kbsearch/internal/store"
)

// QueryEmbedder — минимум, который нужен поиску от модели.
type QueryEmbedder interface {
	EmbedQuery(ctx context.Context, q string) ([]float32, error)
}

// Result — то, что видит пользователь.
type Result struct {
	DocID   string
	Heading string
	Snippet string
	Score   float64
	Cosine  float64
	Lexical float64
}

// Query ищет top-N чанков, оставляя по одному лучшему на документ.
func Query(ctx context.Context, cfg config.Config, st *store.Store, em QueryEmbedder, q string) ([]Result, error) {
	vec, err := em.EmbedQuery(ctx, q)
	if err != nil {
		return nil, err
	}

	hits, err := st.KNN(ctx, vec, cfg.K)
	if err != nil {
		return nil, err
	}

	terms := queryTerms(q)

	if !cfg.Dedupe() {
		out := make([]Result, 0, len(hits))
		for _, h := range hits {
			cos := cosineFromDistance(h.Distance)
			lex := lexicalOverlap(terms, h.Heading+" "+h.Body)
			score := cos
			if cfg.Hybrid() {
				score = (1-cfg.LexicalWeight)*cos + cfg.LexicalWeight*lex
			}
			out = append(out, Result{
				DocID:   h.DocID,
				Heading: h.Heading,
				Snippet: snippet(h.Body, 300),
				Score:   score,
				Cosine:  cos,
				Lexical: lex,
			})
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
		if len(out) > cfg.Top {
			out = out[:cfg.Top]
		}
		return out, nil
	}

	best := make(map[string]Result, len(hits))
	for _, h := range hits {
		cos := cosineFromDistance(h.Distance)
		lex := lexicalOverlap(terms, h.Heading+" "+h.Body)

		score := cos
		if cfg.Hybrid() {
			score = (1-cfg.LexicalWeight)*cos + cfg.LexicalWeight*lex
		}

		cur, ok := best[h.DocID]
		if ok && cur.Score >= score {
			continue
		}
		best[h.DocID] = Result{
			DocID:   h.DocID,
			Heading: h.Heading,
			Snippet: snippet(h.Body, 300),
			Score:   score,
			Cosine:  cos,
			Lexical: lex,
		}
	}

	out := make([]Result, 0, len(best))
	for _, r := range best {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].DocID < out[j].DocID
		}
		return out[i].Score > out[j].Score
	})

	if len(out) > cfg.Top {
		out = out[:cfg.Top]
	}
	return out, nil
}

// cosineFromDistance переводит L2-расстояние в косинус для единичных векторов.
func cosineFromDistance(d float64) float64 {
	c := 1 - d*d/2
	if c < 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return c
}

func queryTerms(q string) []string {
	fields := strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= 'а' && r <= 'я') && r != 'ё' && !(r >= '0' && r <= '9') && r != '-'
	})
	seen := make(map[string]bool, len(fields))
	var out []string
	for _, f := range fields {
		if utf8.RuneCountInString(f) < 4 || seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}

// lexicalOverlap — доля термов запроса, найденных в тексте.
func lexicalOverlap(terms []string, text string) float64 {
	if len(terms) == 0 {
		return 0
	}
	lower := strings.ToLower(text)
	found := 0
	for _, t := range terms {
		if strings.Contains(lower, t) {
			found++
		}
	}
	return float64(found) / float64(len(terms))
}

func snippet(body string, limit int) string {
	body = strings.Join(strings.Fields(body), " ")
	if utf8.RuneCountInString(body) <= limit {
		return body
	}
	return string([]rune(body)[:limit]) + "…"
}
