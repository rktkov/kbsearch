// Команда spike — проверка гейта M0: sqlite-vec загружается и vec0 работает.
// Запуск: go run ./cmd/spike
package main

import (
	"database/sql"
	"fmt"
	"os"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	sqlite_vec.Auto()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fail("open", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE VIRTUAL TABLE vec_chunks USING vec0(embedding float[4])`); err != nil {
		fail("create vec0", err)
	}

	vec := []float32{1, 0, 0, 0}
	blob, err := sqlite_vec.SerializeFloat32(vec)
	if err != nil {
		fail("serialize", err)
	}
	if _, err := db.Exec(`INSERT INTO vec_chunks(rowid, embedding) VALUES (1, ?)`, blob); err != nil {
		fail("insert", err)
	}

	var (
		rowid    int64
		distance float64
	)
	q, err := sqlite_vec.SerializeFloat32([]float32{0.9, 0.1, 0, 0})
	if err != nil {
		fail("serialize query", err)
	}
	err = db.QueryRow(`SELECT rowid, distance FROM vec_chunks WHERE embedding MATCH ? AND k = 1`, q).Scan(&rowid, &distance)
	if err != nil {
		fail("knn", err)
	}

	fmt.Printf("vec0 ok · rowid=%d distance=%.4f\n", rowid, distance)
	fmt.Println("M0 GATE: sqlite-vec PASSED")
}

func fail(stage string, err error) {
	fmt.Fprintf(os.Stderr, "M0 GATE FAILED at %s: %v\n", stage, err)
	os.Exit(1)
}
