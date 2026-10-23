// Package store — SQLite + sqlite-vec: документы, чанки, векторы.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"

	"github.com/rktkov/kbsearch/internal/chunk"
)

// Store — единственный владелец схемы и транзакций.
type Store struct {
	db   *sql.DB
	dims int
}

// Hit — результат KNN до гибридного скоринга.
type Hit struct {
	ChunkID  int64
	DocID    string
	Heading  string
	Body     string
	Distance float64
}

// Open открывает базу и создаёт схему.
func Open(path string, dims int) (*Store, error) {
	sqlite_vec.Auto()

	db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	s := &Store{db: db, dims: dims}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS documents(
			doc_id TEXT PRIMARY KEY,
			path TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			status TEXT NOT NULL,
			embedding_model TEXT NOT NULL,
			indexed_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS chunks(
			id INTEGER PRIMARY KEY,
			doc_id TEXT NOT NULL,
			heading TEXT NOT NULL DEFAULT '',
			body TEXT NOT NULL,
			content_hash TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_doc ON chunks(doc_id)`,
		fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(embedding float[%d])`, s.dims),
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Close закрывает базу.
func (s *Store) Close() error { return s.db.Close() }

// DocHashes возвращает doc_id → content_hash для уже проиндексированных документов.
func (s *Store) DocHashes(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT doc_id, content_hash FROM documents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var id, hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return nil, err
		}
		out[id] = hash
	}
	return out, rows.Err()
}

// UpsertDoc заменяет документ целиком: чанки и векторы — в одной транзакции.
func (s *Store) UpsertDoc(ctx context.Context, docID, path, contentHash, model string, chunks []chunk.Chunk, vecs [][]float32) error {
	if len(chunks) != len(vecs) {
		return fmt.Errorf("чанков %d, векторов %d", len(chunks), len(vecs))
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteDocTx(ctx, tx, docID); err != nil {
		return err
	}

	for i, c := range chunks {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO chunks(doc_id, heading, body, content_hash) VALUES (?, ?, ?, ?)`,
			docID, c.Heading, c.Body, c.Hash)
		if err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		blob, err := sqlite_vec.SerializeFloat32(vecs[i])
		if err != nil {
			return fmt.Errorf("serialize vector: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO vec_chunks(rowid, embedding) VALUES (?, ?)`, id, blob); err != nil {
			return fmt.Errorf("insert vector: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO documents(doc_id, path, content_hash, status, embedding_model, indexed_at)
		VALUES (?, ?, ?, 'indexed', ?, ?)
		ON CONFLICT(doc_id) DO UPDATE SET path=excluded.path, content_hash=excluded.content_hash,
			status=excluded.status, embedding_model=excluded.embedding_model, indexed_at=excluded.indexed_at`,
		docID, path, contentHash, model, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}

	return tx.Commit()
}

// DeleteDoc удаляет документ, его чанки и векторы в одной транзакции.
func (s *Store) DeleteDoc(ctx context.Context, docID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteDocTx(ctx, tx, docID); err != nil {
		return err
	}
	return tx.Commit()
}

func deleteDocTx(ctx context.Context, tx *sql.Tx, docID string) error {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM chunks WHERE doc_id = ?`, docID)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
		args := make([]any, len(ids))
		for i, id := range ids {
			args[i] = id
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM vec_chunks WHERE rowid IN (`+placeholders+`)`, args...); err != nil {
			return fmt.Errorf("delete vectors: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE doc_id = ?`, docID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM documents WHERE doc_id = ?`, docID); err != nil {
		return err
	}
	return nil
}

// KNN ищет ближайшие чанки. k — сколько соседей вернуть.
func (s *Store) KNN(ctx context.Context, query []float32, k int) ([]Hit, error) {
	blob, err := sqlite_vec.SerializeFloat32(query)
	if err != nil {
		return nil, err
	}
	if k < 1 {
		k = 1
	}
	if k > 1000 {
		k = 1000
	}

	// k подставляем числом: vec0 принимает его как параметр плана, не как bind.
	// KNN выносим в подзапрос: vec0 не любит MATCH в join-плане.
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT c.id, c.doc_id, c.heading, c.body, knn.distance
		FROM (SELECT rowid, distance FROM vec_chunks WHERE embedding MATCH ? AND k = %d) AS knn
		JOIN chunks c ON c.id = knn.rowid
		ORDER BY knn.distance`, k), blob)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.ChunkID, &h.DocID, &h.Heading, &h.Body, &h.Distance); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// Stats возвращает число документов и чанков.
func (s *Store) Stats(ctx context.Context) (docs, chunks int, err error) {
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&docs); err != nil {
		return 0, 0, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&chunks); err != nil {
		return 0, 0, err
	}
	return docs, chunks, nil
}
