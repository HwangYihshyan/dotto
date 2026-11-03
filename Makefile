# よく使うコマンドの短縮形。 make up / make test のように使う。
# （Windows で make が無い場合は、各行のコマンドをそのまま打てばよい）

.PHONY: up down logs reset test psql redis

up:        ## 全サービスをビルドして起動
	docker compose up -d --build

down:      ## 停止（データは残る）
	docker compose down

logs:      ## API のログを追いかける
	docker compose logs -f api

reset:     ## 停止してデータも消す（初期化SQLをやり直したいとき）
	docker compose down -v

test:      ## テストを実行
	go test ./... -race -count=1

psql:      ## PostgreSQL に入る
	docker compose exec db psql -U dotto -d dotto

redis:     ## Redis に入る
	docker compose exec redis redis-cli
