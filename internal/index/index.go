// Package index синхронизирует корпус с хранилищем: добавить, обновить, удалить.
package index

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/rktkov/kbsearch/internal/chunk"
	"github.com/rktkov/kbsearch/internal/config"
	"github.com/rktkov/kbsearch/internal/corpus"
	"github.com/rktkov/kbsearch/internal/store"
)

// Embedder — то, что нужно индексатору от модели. Интерфейс, чтобы тесты шли без Ollama.
type Embedder interface {
	EmbedPassages(ctx context.Context, texts []string) ([][]float32, error)
	Model() string
	Dims() int
}

// Report — что изменилось за один прогон.
type Report struct {
	Added, Updated, Deleted, Unchanged int
}

func (r Report) String() string {
	return fmt.Sprintf("added=%d updated=%d deleted=%d unchanged=%d",
		r.Added, r.Updated, r.Deleted, r.Unchanged)
}

// Sync приводит индекс в соответствие с каталогом документов.
//
// Инвариант: после успешного Sync для каждого файла в docsDir есть ровно один
// документ в БД, а для удалённых файлов не осталось ни чанков, ни векторов.
func Sync(ctx context.Context, docsDir string, st *store.Store, em Embedder, cfg config.Config) (Report, error) {
	docs, err := corpus.Load(docsDir)
	if err != nil {
		return Report{}, err
	}
	existing, err := st.DocHashes(ctx)
	if err != nil {
		return Report{}, err
	}

	var rep Report
	seen := make(map[string]bool, len(docs))

	for _, doc := range docs {
		seen[doc.ID] = true
		hash := contentHash(doc.Content)

		if prev, ok := existing[doc.ID]; ok && prev == hash {
			rep.Unchanged++
			continue
		}

		chunks := chunk.Document(doc.ID, doc.Content, cfg.MaxChunkRunes, cfg.HeadingAware())
		if len(chunks) == 0 {
			continue
		}

		texts := make([]string, len(chunks))
		for i, c := range chunks {
			texts[i] = c.Text
		}
		vecs, err := em.EmbedPassages(ctx, texts)
		if err != nil {
			return rep, fmt.Errorf("embed %s: %w", doc.ID, err)
		}

		if err := st.UpsertDoc(ctx, doc.ID, doc.Path, hash, em.Model(), chunks, vecs); err != nil {
			return rep, fmt.Errorf("upsert %s: %w", doc.ID, err)
		}

		if _, ok := existing[doc.ID]; ok {
			rep.Updated++
		} else {
			rep.Added++
		}
	}

	for docID := range existing {
		if seen[docID] {
			continue
		}
		if err := st.DeleteDoc(ctx, docID); err != nil {
			return rep, fmt.Errorf("delete %s: %w", docID, err)
		}
		rep.Deleted++
	}

	return rep, nil
}

func contentHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
