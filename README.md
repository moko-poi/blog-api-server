# ブログサービス

Mat Ryer氏の13年間のGoでのHTTPサービス開発のベストプラクティスに従って構築された、本番運用対応のブログ投稿管理HTTPサービスです。

## 特徴

- **RESTful API** によるブログのCRUD操作
- **構造化ログ** と設定可能なログレベル
- **グレースフルシャットダウン** とシグナルハンドリング
- **包括的テスト** （統合テストを含む）
- **入力バリデーション** と詳細なエラーメッセージ
- **ミドルウェア対応** （ログ出力、CORS、パニック回復、レート制限）
- **ヘルスチェック** （モニタリングと準備完了プローブ）
- **Docker対応** （マルチステージビルド）
- **本番対応** の設定管理

## Mat Ryerのベストプラクティス実装ポイント

### 1. NewServerパターン
```go
func NewServer(log *logger.Logger, cfg *config.Config, blogStore store.BlogStore) (*Server, error)
```
- すべての依存関係を引数として明示的に受け取る
- 依存関係の注入により、テスタビリティと保守性を向上

### 2. run関数パターン
```go
func run(ctx context.Context, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer, args []string) error
```
- OSの基本要素（環境変数、標準入出力、引数）を注入可能
- テスト時にこれらを制御でき、並列テスト実行が可能

### 3. ルート定義の一元化
- `routes.go`でAPI全体の構造を一箇所に定義
- サービスのAPI表面が一目でわかり、メンテナンスが容易

### 4. maker funcパターン
```go
func handleBlogsCreate(log *logger.Logger, blogStore store.BlogStore) http.Handler
```
- ハンドラー関数は依存関係を受け取り、`http.Handler`を返す
- クロージャ環境で初期化処理を実行可能

### 5. encode/decodeの一元化
- JSONのエンコード/デコード処理を一箇所で実装
- 将来的な形式変更（XML対応等）への対応が容易

### 6. バリデーションインターフェース
```go
type Validator interface {
    Valid(ctx context.Context) (problems map[string]string)
}
```
- シンプルな単一メソッドインターフェース
- オブジェクト自身がバリデーション責任を持つ

## APIエンドポイント

### ヘルスチェック
- `GET /healthz` - ヘルスチェック
- `GET /readyz` - 準備完了チェック

### ブログ管理
- `GET /api/v1/blogs` - 全ブログ一覧取得
- `GET /api/v1/blogs?author=<name>` - 作者でフィルタリング
- `POST /api/v1/blogs` - 新規ブログ作成
- `GET /api/v1/blogs/{id}` - 特定ブログ取得
- `PUT /api/v1/blogs/{id}` - ブログ更新
- `DELETE /api/v1/blogs/{id}` - ブログ削除

## プロジェクト構成

```
blog-service/
├── cmd/
│   └── server/
│       └── main.go              # アプリケーションエントリーポイント
├── internal/
│   ├── api/
│   │   ├── handlers.go          # HTTPハンドラー
│   │   ├── middleware.go        # HTTPミドルウェア
│   │   ├── routes.go            # ルート定義
│   │   ├── server.go            # サーバー設定とライフサイクル
│   │   └── validation.go        # リクエスト/レスポンスバリデーション
│   ├── config/
│   │   └── config.go            # 設定管理
│   ├── logger/
│   │   └── logger.go            # 構造化ログ
│   └── store/
│       ├── memory.go            # インメモリストレージ実装
│       └── store.go             # ストレージインターフェース
├── pkg/
│   └── models/
│       └── blog.go              # ドメインモデル
├── tests/
│   └── integration/
│       └── api_test.go          # 統合テスト
├── deployments/
│   ├── docker/
│   │   └── Dockerfile           # Dockerビルド設定
│   └── k8s/                     # Kubernetesマニフェスト（将来実装）
├── go.mod
├── go.sum
├── Makefile                     # ビルドと開発タスク
└── README.md
```

## 開始方法

### 前提条件

- Go 1.21以降
- Make（Makefileターゲット使用のため、オプション）
- Docker（コンテナ化のため、オプション）

### ローカル開発

1. **リポジトリのクローン**
   ```bash
   git clone <repository-url>
   cd blog-service
   ```

2. **依存関係のインストール**
   ```bash
   make deps
   ```

3. **テスト実行**
   ```bash
   make test
   make test-integration
   ```

4. **サービス起動**
   ```bash
   make run-dev
   ```

   サービスは `http://localhost:8080` で起動します

### 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|---# Blog Service

A production-ready HTTP service for managing blog posts, built following Mat Ryer's 13 years of Go HTTP service best practices.

## Features

- **RESTful API** for blog CRUD operations
- **Structured logging** with configurable levels
- **Graceful shutdown** with proper signal handling
- **Comprehensive testing** including integration tests
- **Input validation** with detailed error messages
- **Middleware support** (logging, CORS, panic recovery, rate limiting)
- **Health checks** for monitoring and readiness probes
- **Docker support** with multi-stage builds
- **Production-ready** configuration management

## API Endpoints

### Health Checks
- `GET /healthz` - Health check endpoint
- `GET /readyz` - Readiness check endpoint

### Blog Management
- `GET /api/v1/blogs` - List all blogs
- `GET /api/v1/blogs?author=<name>` - Filter blogs by author
- `POST /api/v1/blogs` - Create a new blog
- `GET /api/v1/blogs/{id}` - Get a specific blog
- `PUT /api/v1/blogs/{id}` - Update a blog
- `DELETE /api/v1/blogs/{id}` - Delete a blog

## Project Structure

```
blog-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go          # HTTP handlers
│   │   ├── middleware.go        # HTTP middleware
│   │   ├── routes.go            # Route definitions
│   │   ├── server.go            # Server setup and lifecycle
│   │   └── validation.go        # Request/response validation
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   └── store/
│       ├── memory.go            # In-memory storage implementation
│       └── store.go             # Storage interface
├── pkg/
│   └── models/
│       └── blog.go              # Domain models
├── tests/
│   └── integration/
│       └── api_test.go          # Integration tests
├── deployments/
│   ├── docker/
│   │   └── Dockerfile           # Docker build configuration
│   └── k8s/                     # Kubernetes manifests (future)
├── go.mod
├── go.sum
├── Makefile                     # Build and development tasks
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile targets)
- Docker (optional, for containerization)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd blog-service
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Run tests**
   ```bash
   make test
   make test-integration
   ```

4. **Run the service**
   ```bash
   make run-dev
   ```

   The service will start on `http://localhost:8080`

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | Server host |
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `READ_TIMEOUT` | `10` | HTTP read timeout in seconds |
|
