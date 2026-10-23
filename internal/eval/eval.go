// Package eval — набор запросов с ожидаемыми документами и метрики качества.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rktkov/kbsearch/internal/config"
	"github.com/rktkov/kbsearch/internal/search"
	"github.com/rktkov/kbsearch/internal/store"
)

// Case — один запрос и документы, которые считаются правильным ответом.
type Case struct {
	Query  string   `json:"q"`
	Expect []string `json:"expect"`
}

// Metrics — итог прогона.
type Metrics struct {
	N        int
	RecallAt3 float64
	MRR       float64
	P95ms     float64
}

// Row — строка отчёта по одному запросу.
type Row struct {
	Query  string
	Expect []string
	Got    []string
	Rank   int // позиция первого правильного (0 = не найден)
	MS     float64
}

// Embedder — то, что нужно от модели.
type Embedder interface {
	EmbedQuery(ctx context.Context, q string) ([]float32, error)
}

// Run прогоняет набор и считает recall@3, MRR и p95.
func Run(ctx context.Context, cfg config.Config, st *store.Store, em Embedder, path string) (Metrics, []Row, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Metrics{}, nil, fmt.Errorf("read queries: %w", err)
	}
	var cases []Case
	if err := json.Unmarshal(raw, &cases); err != nil {
		return Metrics{}, nil, fmt.Errorf("parse queries: %w", err)
	}
	if len(cases) == 0 {
		return Metrics{}, nil, fmt.Errorf("пустой набор запросов")
	}

	var (
		rows    []Row
		recall  int
		mrr     float64
		latency []float64
	)

	for _, c := range cases {
		start := time.Now()
		results, err := search.Query(ctx, cfg, st, em, c.Query)
		if err != nil {
			return Metrics{}, nil, fmt.Errorf("query %q: %w", c.Query, err)
		}
		ms := float64(time.Since(start).Microseconds()) / 1000
		latency = append(latency, ms)

		row := Row{Query: c.Query, Expect: c.Expect, MS: ms}
		for i, r := range results {
			row.Got = append(row.Got, r.DocID)
			if row.Rank == 0 && contains(c.Expect, r.DocID) {
				row.Rank = i + 1
			}
		}

		if row.Rank > 0 && row.Rank <= 3 {
			recall++
		}
		if row.Rank > 0 {
			mrr += 1 / float64(row.Rank)
		}
		rows = append(rows, row)
	}

	sort.Float64s(latency)
	p95 := percentile(latency, 0.95)

	return Metrics{
		N:         len(cases),
		RecallAt3: float64(recall) / float64(len(cases)),
		MRR:       mrr / float64(len(cases)),
		P95ms:     p95,
	}, rows, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}
