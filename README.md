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
- **現代的開発ツール** （golangci-lint、lefthook、asdf）
- **Git hooks** による自動品質チェック

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
blog-api-server/
├── cmd/
│   └── server/
│       └── main.go              # アプリケーションエントリーポイント
├── internal/
│   ├── api/
│   │   ├── handlers.go          # HTTPハンドラー
│   │   ├── handlers_test.go     # ハンドラーテスト
│   │   ├── middleware.go        # HTTPミドルウェア
│   │   ├── middleware_test.go   # ミドルウェアテスト
│   │   ├── routes.go            # ルート定義
│   │   ├── routes_test.go       # ルートテスト
│   │   ├── server.go            # サーバー設定とライフサイクル
│   │   ├── validation.go        # リクエスト/レスポンスバリデーション
│   │   └── validation_test.go   # バリデーションテスト
│   ├── config/
│   │   └── config.go            # 設定管理
│   ├── domain/
│   │   ├── blog.go              # ドメインモデル
│   │   └── blog_test.go         # ドメインモデルテスト
│   ├── logger/
│   │   └── logger.go            # 構造化ログ
│   └── store/
│       ├── store.go             # ストレージインターフェース
│       └── store_test.go        # ストレージテスト
├── scripts/
│   ├── setup.sh                 # 開発環境セットアップスクリプト
│   ├── dev.sh                   # 開発サーバー起動スクリプト
│   ├── test.sh                  # テスト実行スクリプト
│   └── build.sh                 # ビルドスクリプト
├── .vscode/                     # VS Code設定
├── .tool-versions               # asdf ツールバージョン定義
├── .golangci.yml                # golangci-lint設定
├── lefthook.yml                 # Git hooks設定
├── Dockerfile                   # 本番用Dockerイメージ
├── Dockerfile.dev               # 開発用Dockerイメージ
├── docker-compose.dev.yml       # 開発用Docker Compose
├── .air.toml                    # ホットリロード設定
├── .env                         # 環境変数
├── .env.example                 # 環境変数テンプレート
├── .gitignore                   # Git除外設定
├── go.mod
├── go.sum
├── Makefile                     # ビルドと開発タスク
├── DEVELOPMENT.md               # 開発者向けドキュメント
└── README.md
```

## 開始方法

### 前提条件

- Go 1.24.4以降
- Make（Makefileターゲット使用のため、オプション）
- Docker（コンテナ化のため、オプション）
- asdf（ツールバージョン管理のため、推奨）

### ローカル開発

1. **リポジトリのクローン**
   ```bash
   git clone <repository-url>
   cd blog-api-server
   ```

2. **ツールのインストール（推奨）**
   ```bash
   # asdfを使用する場合（推奨）
   asdf install

   # または手動でセットアップ
   make setup
   ```

3. **テスト実行**
   ```bash
   make test
   make test-cover
   ```

4. **サービス起動（ホットリロード付き）**
   ```bash
   make dev
   ```

   サービスは `http://localhost:8080` で起動します

### Docker開発環境

```bash
# 開発環境を起動（ホットリロード付き）
docker compose -f docker-compose.dev.yml up

# バックグラウンドで起動
docker compose -f docker-compose.dev.yml up -d
```

### 環境変数

| 変数名 | デフォルト | 説明 |
|--------|-----------|------|
| `HOST` | `localhost` | サーバーホスト |
| `PORT` | `8080` | サーバーポート |
| `LOG_LEVEL` | `debug` | ログレベル (debug, info, warn, error) |
| `READ_TIMEOUT` | `10s` | HTTP読み取りタイムアウト |
| `WRITE_TIMEOUT` | `10s` | HTTP書き込みタイムアウト |
| `IDLE_TIMEOUT` | `120s` | HTTPアイドルタイムアウト |
| `DEV_MODE` | `true` | 開発モード |

詳細は `.env.example` を参照してください。

## 🛠️ 利用可能なコマンド

### Make コマンド

| コマンド | 説明 |
|----------|------|
| `make help` | 利用可能なコマンド一覧を表示 |
| `make setup` | 初期開発環境セットアップ |
| `make dev` | ホットリロード付きで開発サーバーを起動 |
| `make build` | 本番用バイナリをビルド |
| `make build-all` | 複数プラットフォーム向けにビルド |
| `make test` | 全テストを実行 |
| `make test-cover` | カバレッジ付きでテストを実行 |
| `make lint` | golangci-lintでコード解析 |
| `make lint-fix` | golangci-lintで自動修正 |
| `make audit` | 包括的なコード品質監査 |
| `make hooks-run` | Git hooksを手動実行 |
| `make clean` | ビルド成果物をクリーンアップ |

## 🔧 開発ツール

このプロジェクトでは現代的な開発ツールを使用してコード品質と開発効率を向上させています。

### asdf ツールバージョン管理

`.tool-versions` ファイルで開発ツールのバージョンを統一管理：

```bash
# asdfでツールをインストール
asdf install

# 個別ツールのインストール
asdf install golang 1.24.4
asdf install golangci-lint 1.62.2
```

### golangci-lint 包括的リンター

50以上のリンターでコード品質をチェック：

```bash
# リント実行
make lint

# 自動修正付きリント
make lint-fix

# 設定ファイル: .golangci.yml
```

### lefthook Git hooks管理

Git操作時に自動でコード品質チェックを実行：

```bash
# Git hooks設定
make setup-hooks

# hooks手動実行
make hooks-run

# 設定ファイル: lefthook.yml
```

**自動実行されるチェック:**
- コミット時: フォーマット、テスト、リント
- プッシュ時: 包括テスト、セキュリティ監査
- コミットメッセージ: 規約チェック

詳細な開発ガイドは [DEVELOPMENT.md](DEVELOPMENT.md) を参照してください。
