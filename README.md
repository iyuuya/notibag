# Notibag

オレ用通知システム。

## Dev

```bash
# バックエンド開発
cd backend
go run main.go

# フロントエンド開発
cd frontend
npm run dev

# Docker環境
docker compose up --build
```

## CLI コマンド

### ビルド

```bash
cd backend
make build
```

### 使用方法

```bash
./notibag-send -title "通知タイトル" -message "通知メッセージ"
```

### オプション

- `-host`: サーバーホストURL (デフォルト: 設定ファイルから読み込み)
- `-title`: 通知タイトル (必須)
- `-message`: 通知メッセージ (必須)

### 設定ファイル

`~/.notibag/config.json` でデフォルトのホストを設定できます。

```json
{
  "host": "http://localhost:8080"
}
```

## プロジェクト構造

```
.
├── backend/          # Go WebSocket/API サーバー
│   ├── cmd/         # CLI コマンド
│   └── main.go      # サーバー
├── frontend/         # Vite + React アプリ
├── nginx/           # Nginx 設定
├── docker-compose.yml
└── README.md
```
