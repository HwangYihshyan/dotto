# Dotto — ステップ1：ローカル環境

Go API・PostgreSQL・Redis を Docker Compose でまとめて起動する土台です。

## 必要なもの
- Docker Desktop（Windows の場合は WSL2 を有効にする）
- Go（最新版）
- Git

## 初回セットアップ
```bash
cp .env.example .env      # 環境変数ファイルを作る（パスワードは自由に変えてよい）
go mod tidy               # 依存ライブラリ（pgx, go-redis）を取得し go.sum を作る
make up                   # ビルドして起動（make が無ければ: docker compose up -d --build）
```

## 動作確認
```bash
curl localhost:8080/healthz   # {"status":"ok"}
curl localhost:8080/readyz    # {"postgres":"ok","redis":"ok"}
docker compose ps             # 3つとも healthy / running になっていること
```

## よく使うコマンド
| コマンド | 内容 |
|---|---|
| `make up` | ビルドして起動 |
| `make logs` | API のログを見る |
| `make test` | テスト実行 |
| `make psql` | DB に入る（`\dt` でテーブル一覧） |
| `make down` | 停止（データは残る） |
| `make reset` | 停止してデータも削除 |

## ディレクトリ構成
```
cmd/api/main.go           入口：設定読み込み → DB/Redis 接続 → サーバー起動・終了処理
internal/config/          環境変数の読み込み
internal/server/          ルーティング・ハンドラ・ミドルウェア・テスト
db/init/001_init.sql      DB 初回起動時に流れるテーブル定義
Dockerfile                マルチステージビルド
compose.yaml              api / db / redis の3サービス定義
```
