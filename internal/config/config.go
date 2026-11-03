// Package config は環境変数からアプリの設定を読み込む。
// 設定値をコードに直書きせず環境変数で渡すことで、
// ローカル・CI・本番で同じイメージを使い回せる（12-Factor App の考え方）。
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string // API が待ち受けるポート
	DatabaseURL string // PostgreSQL の接続文字列
	RedisAddr   string // Redis のアドレス（host:port）
}

// Load は環境変数を読み込む。必須の値が無ければエラーを返す。
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL が設定されていません")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
