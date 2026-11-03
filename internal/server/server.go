// Package server は HTTP のルーティングとハンドラをまとめる。
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger は「接続先が生きているか確認できるもの」を表すインターフェース。
// PostgreSQL や Redis を直接受け取らずにインターフェースにしておくと、
// テストでは偽物（モック）を渡せる。
type Pinger interface {
	Ping(ctx context.Context) error
}

type Server struct {
	db    Pinger
	redis Pinger
}

func New(db, redis Pinger) *Server {
	return &Server{db: db, redis: redis}
}

// Routes はルーティングを登録した http.Handler を返す。
// Go 1.22 以降の標準ライブラリは "GET /path" のようにメソッド指定ができる。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	return logRequests(mux)
}

// /healthz : プロセスが動いているかだけを返す（liveness）。
// DB が落ちていても 200 を返す。ここが落ちたら「アプリを再起動すべき」という意味。
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// /readyz : DB と Redis に繋がるかまで確認する（readiness）。
// ここが落ちたら「今はリクエストを受けられない」という意味。
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{"postgres": "ok", "redis": "ok"}
	status := http.StatusOK

	if err := s.db.Ping(ctx); err != nil {
		checks["postgres"] = err.Error()
		status = http.StatusServiceUnavailable
	}
	if err := s.redis.Ping(ctx); err != nil {
		checks["redis"] = err.Error()
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, checks)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
