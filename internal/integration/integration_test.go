//go:build integration

// Package integration は本物の PostgreSQL / Redis に接続して確かめる統合テスト。
//
// 先頭の "//go:build integration" により、普段の `go test ./...` では実行されない。
// 実行するときは DB と Redis を起動した状態で:
//
//	go test -tags=integration ./internal/integration/... -v
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func newDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL が未設定なのでスキップ")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("DB 接続に失敗: %v", err)
	}
	t.Cleanup(db.Close) // テスト終了時に必ず閉じる
	return db
}

func TestCanvasesTable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := newDB(t)

	// トランザクション内で操作し、最後に Rollback するので DB にデータが残らない
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	t.Run("正しい値なら INSERT できる", func(t *testing.T) {
		var id int64
		err := tx.QueryRow(ctx,
			`INSERT INTO canvases (name, width, height) VALUES ($1, $2, $3) RETURNING id`,
			"test", 32, 32,
		).Scan(&id)
		if err != nil {
			t.Fatalf("INSERT に失敗: %v", err)
		}
		if id <= 0 {
			t.Fatalf("id = %d, want > 0", id)
		}
	})

	t.Run("幅0は CHECK 制約で弾かれる", func(t *testing.T) {
		// SAVEPOINT：失敗してもトランザクション全体を壊さないための区切り
		if _, err := tx.Exec(ctx, `SAVEPOINT sp`); err != nil {
			t.Fatal(err)
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO canvases (name, width, height) VALUES ('bad', 0, 32)`)
		if err == nil {
			t.Fatal("エラーになるはずが成功した")
		}
		if _, err := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT sp`); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRedisSetGet(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR が未設定なのでスキップ")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })

	key := "integration-test:" + t.Name()
	t.Cleanup(func() { rdb.Del(context.Background(), key) })

	if err := rdb.Set(ctx, key, "hello", time.Minute).Err(); err != nil {
		t.Fatalf("SET に失敗: %v", err)
	}
	got, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("GET に失敗: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}
