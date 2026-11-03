// main はアプリの入口。設定の読み込み → 外部サービスへの接続 → サーバー起動 を行う。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/HwangYihshyan/dotto/internal/config"
	"github.com/HwangYihshyan/dotto/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("起動に失敗しました", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Ctrl+C や docker stop（SIGTERM）を受け取ったら ctx がキャンセルされる
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// PostgreSQL: 接続プールを作る（リクエストごとに接続を作り直さないため）
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(db, redisPinger{rdb}).Routes(),
		ReadHeaderTimeout: 5 * time.Second, // 遅いクライアントによる攻撃（Slowloris）対策
	}

	// サーバーは別の goroutine で動かし、main はシグナルを待つ
	errCh := make(chan error, 1)
	go func() {
		slog.Info("サーバーを起動しました", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// グレースフルシャットダウン：処理中のリクエストが終わるのを最大10秒待つ
	slog.Info("シャットダウンしています")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// redisPinger は go-redis のクライアントを server.Pinger に合わせるためのアダプタ。
type redisPinger struct{ c *redis.Client }

func (p redisPinger) Ping(ctx context.Context) error { return p.c.Ping(ctx).Err() }
