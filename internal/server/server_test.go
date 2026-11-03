package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakePinger はテスト用の偽物。err を返すかどうかを自由に決められる。
type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealthz(t *testing.T) {
	s := New(fakePinger{}, fakePinger{})
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// テーブル駆動テスト：Go でよく使われる書き方。ケースを足すのが簡単。
func TestReadyz(t *testing.T) {
	down := errors.New("connection refused")

	tests := []struct {
		name  string
		db    error
		redis error
		want  int
	}{
		{"両方OK", nil, nil, http.StatusOK},
		{"DBが落ちている", down, nil, http.StatusServiceUnavailable},
		{"Redisが落ちている", nil, down, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(fakePinger{tt.db}, fakePinger{tt.redis})
			rec := httptest.NewRecorder()
			s.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
