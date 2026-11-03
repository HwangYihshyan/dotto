-- PostgreSQL コンテナの「初回起動時だけ」実行される初期化 SQL。
-- （データ用ボリュームが空のときだけ動く。テーブルを変えたら `make reset` で作り直す）
-- 本格的なマイグレーションはステップ2以降で golang-migrate などに置き換える。

CREATE TABLE IF NOT EXISTS canvases (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    width      INT         NOT NULL CHECK (width  BETWEEN 1 AND 256),
    height     INT         NOT NULL CHECK (height BETWEEN 1 AND 256),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
