package index

import (
	"context"
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"

	"github.com/rktkov/kbsearch/internal/config"
	"github.com/rktkov/kbsearch/internal/store"
)

// fakeEmbedder даёт детерминированные векторы без Ollama.
type fakeEmbedder struct{ dims int }

func (f fakeEmbedder) Model() string { return "fake" }
func (f fakeEmbedder) Dims() int     { return f.dims }

func (f fakeEmbedder) EmbedPassages(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		sum := sha1.Sum([]byte(t))
		v := make([]float32, f.dims)
		for j := range v {
			v[j] = float32(sum[j]) / 255
		}
		out[i] = v
	}
	return out, nil
}

func testConfig(t *testing.T, dir string) config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.DocsDir = dir
	cfg.MaxChunkRunes = 120
	cfg.Profile = config.Fixed
	return cfg
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"), 4)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func writeDoc(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write doc: %v", err)
	}
}

func TestSyncAddsUpdatesAndDeletes(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeDoc(t, dir, "a.md", "# A\n\nПервый документ про инварианты и агрегаты.\n")
	writeDoc(t, dir, "b.md", "# B\n\nВторой документ про потоки данных.\n")

	st := openStore(t)
	cfg := testConfig(t, dir)
	em := fakeEmbedder{dims: 4}

	rep, err := Sync(ctx, dir, st, em, cfg)
	if err != nil {
		t.Fatalf("sync 1: %v", err)
	}
	if rep.Added != 2 || rep.Updated != 0 || rep.Deleted != 0 {
		t.Fatalf("ожидали added=2, получили %s", rep)
	}

	rep, err = Sync(ctx, dir, st, em, cfg)
	if err != nil {
		t.Fatalf("sync 2: %v", err)
	}
	if rep.Unchanged != 2 || rep.Added != 0 {
		t.Fatalf("повторный sync должен быть no-op, получили %s", rep)
	}

	writeDoc(t, dir, "a.md", "# A\n\nИзменённый документ про инварианты.\n")
	rep, err = Sync(ctx, dir, st, em, cfg)
	if err != nil {
		t.Fatalf("sync 3: %v", err)
	}
	if rep.Updated != 1 {
		t.Fatalf("ожидали updated=1, получили %s", rep)
	}

	if err := os.Remove(filepath.Join(dir, "b.md")); err != nil {
		t.Fatal(err)
	}
	rep, err = Sync(ctx, dir, st, em, cfg)
	if err != nil {
		t.Fatalf("sync 4: %v", err)
	}
	if rep.Deleted != 1 {
		t.Fatalf("ожидали deleted=1, получили %s", rep)
	}

	docs, chunks, err := st.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if docs != 1 {
		t.Fatalf("ожидали 1 документ, получили %d", docs)
	}
	hits, err := st.KNN(ctx, []float32{1, 0, 0, 0}, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hits {
		if h.DocID == "b" {
			t.Fatalf("чанки удалённого документа остались в индексе (chunks=%d)", chunks)
		}
	}
}

func TestSyncSkipsEmptyCorpus(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st := openStore(t)

	rep, err := Sync(ctx, dir, st, fakeEmbedder{dims: 4}, testConfig(t, dir))
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if rep != (Report{}) {
		t.Fatalf("на пустом корпусе ожидали нулевой отчёт, получили %s", rep)
	}
}
