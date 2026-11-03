# ---- ビルド用ステージ：Go のコンパイラが入った大きいイメージ ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# 依存関係だけ先にコピーしてダウンロード。
# go.mod が変わらない限りこの層はキャッシュされ、2回目以降のビルドが速くなる。
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 で外部ライブラリに依存しない単体バイナリを作る
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- 実行用ステージ：バイナリだけ入った小さいイメージ ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
# nonroot イメージなので root 以外のユーザーで動く（セキュリティ上の基本）
ENTRYPOINT ["/api"]
