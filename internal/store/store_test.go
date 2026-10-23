package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rktkov/kbsearch/internal/chunk"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.sqlite"), 4)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestUpsertAndDeleteKeepVectorsInSync(t *testing.T) {
	ctx := context.Background()
	st := openTestStore(t)

	chunks := []chunk.Chunk{
		{DocID: "doc", Heading: "A", Body: "первый", Hash: "h1"},
		{DocID: "doc", Heading: "B", Body: "второй", Hash: "h2"},
	}
	vecs := [][]float32{{1, 0, 0, 0}, {0, 1, 0, 0}}

	if err := st.UpsertDoc(ctx, "doc", "doc.md", "hash1", "fake", chunks, vecs); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	docs, count, err := st.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if docs != 1 || count != 2 {
		t.Fatalf("ожидали 1 документ и 2 чанка, получили %d и %d", docs, count)
	}

	// Повторный upsert не должен плодить дубли.
	if err := st.UpsertDoc(ctx, "doc", "doc.md", "hash1", "fake", chunks, vecs); err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	_, count, _ = st.Stats(ctx)
	if count != 2 {
		t.Fatalf("после повторного upsert ожидали 2 чанка, получили %d", count)
	}

	if err := st.DeleteDoc(ctx, "doc"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	docs, count, _ = st.Stats(ctx)
	if docs != 0 || count != 0 {
		t.Fatalf("после удаления ожидали пусто, получили %d/%d", docs, count)
	}

	hits, err := st.KNN(ctx, []float32{1, 0, 0, 0}, 5)
	if err != nil {
		t.Fatalf("knn: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("векторы удалённого документа остались: %d хитов", len(hits))
	}
}

func TestKNNFindsNearest(t *testing.T) {
	ctx := context.Background()
	st := openTestStore(t)

	chunks := []chunk.Chunk{
		{DocID: "near", Heading: "", Body: "близкий", Hash: "n"},
		{DocID: "far", Heading: "", Body: "далёкий", Hash: "f"},
	}
	vecs := [][]float32{{1, 0, 0, 0}, {0, 1, 0, 0}}
	if err := st.UpsertDoc(ctx, "near", "near.md", "h1", "fake", chunks[:1], vecs[:1]); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertDoc(ctx, "far", "far.md", "h2", "fake", chunks[1:], vecs[1:]); err != nil {
		t.Fatal(err)
	}

	hits, err := st.KNN(ctx, []float32{0.9, 0.1, 0, 0}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("ожидали 2 хита, получили %d", len(hits))
	}
	if hits[0].DocID != "near" {
		t.Fatalf("ближайший должен быть near, получили %s", hits[0].DocID)
	}
}
