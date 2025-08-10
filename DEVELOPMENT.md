# Development Guide

このドキュメントではブログAPIサーバーのローカル開発環境のセットアップと開発ワークフローについて説明します。

## 🚀 クイックスタート

### 前提条件

- Go 1.24.4以降
- Git
- Make（オプション）
- Docker & Docker Compose（オプション）

### 初期セットアップ

```bash
# リポジトリをクローン
git clone <repository-url>
cd blog-api-server

# 開発環境のセットアップ
./scripts/setup.sh
# または
make setup
```

### 開発サーバーの起動

```bash
# ホットリロード付きで開発サーバーを起動
make dev
# または
./scripts/dev.sh
```

サーバーは `http://localhost:8080` で起動します。

## 📋 利用可能なコマンド

### Make コマンド

| コマンド | 説明 |
|----------|------|
| `make help` | 利用可能なコマンド一覧を表示 |
| `make dev` | ホットリロード付きで開発サーバーを起動 |
| `make build` | 本番用バイナリをビルド |
| `make build-all` | 複数プラットフォーム向けにビルド |
| `make test` | 全テストを実行 |
| `make test-cover` | カバレッジ付きでテストを実行 |
| `make test-integration` | 統合テストを実行 |
| `make fmt` | コードフォーマット |
| `make vet` | go vet を実行 |
| `make lint` | 静的解析を実行 |
| `make audit` | 包括的なコード品質監査 |
| `make deps` | 依存関係の管理 |
| `make clean` | ビルド成果物をクリーンアップ |

### スクリプト

| スクリプト | 説明 |
|-----------|------|
| `./scripts/setup.sh` | 初期開発環境セットアップ |
| `./scripts/dev.sh` | 開発サーバー起動 |
| `./scripts/test.sh` | 包括的テスト実行 |
| `./scripts/build.sh` | 本番ビルド |

## 🐳 Docker開発環境

### Docker Compose での開発

```bash
# 開発環境を起動（ホットリロード付き）
docker-compose -f docker-compose.dev.yml up

# バックグラウンドで起動
docker-compose -f docker-compose.dev.yml up -d

# 停止
docker-compose -f docker-compose.dev.yml down
```

### 単体Docker

```bash
# 本番用イメージをビルド
make docker-build

# コンテナを実行
make docker-run
```

## 🧪 テスト

### テストの種類

- **単体テスト**: 個別関数・メソッドのテスト
- **統合テスト**: APIエンドポイントの結合テスト
- **ベンチマークテスト**: パフォーマンステスト

### テストの実行

```bash
# 全テスト実行
make test

# カバレッジ付きテスト
make test-cover

# 統合テストのみ
make test-integration

# ベンチマーク付き包括テスト
./scripts/test.sh
```

### テストファイルの命名規則

- 単体テスト: `*_test.go`
- 統合テスト: `*_integration_test.go` または `// +build integration`タグ付き
- ベンチマーク: `func BenchmarkXxx(*testing.B)`

## 🔧 コード品質

### 自動チェック

プロジェクトでは以下のツールを使用してコード品質を保証しています：

- **gofmt**: コードフォーマット
- **go vet**: 静的解析
- **staticcheck**: 高度な静的解析
- **govulncheck**: 脆弱性チェック

### 品質チェックの実行

```bash
# 包括的な品質監査
make audit

# 個別チェック
make fmt      # フォーマット
make vet      # go vet
make lint     # staticcheck
make vuln     # 脆弱性チェック
```

## 🏗️ プロジェクト構造

```
blog-api-server/
├── cmd/server/           # アプリケーションエントリーポイント
├── internal/             # プライベートパッケージ
│   ├── api/             # HTTP API実装
│   ├── config/          # 設定管理
│   ├── domain/          # ドメインモデル
│   ├── logger/          # ログ機能
│   └── store/           # データストレージ
├── scripts/             # 開発・運用スクリプト
├── build/               # ビルド成果物（gitignore済み）
├── tmp/                 # 一時ファイル（gitignore済み）
├── .vscode/             # VS Code設定
├── Makefile             # ビルドタスク定義
├── Dockerfile           # 本番用コンテナ
├── Dockerfile.dev       # 開発用コンテナ
├── docker-compose.dev.yml # 開発用compose
├── .air.toml            # ホットリロード設定
├── .env                 # 環境変数
└── .env.example         # 環境変数テンプレート
```

## 🌐 環境変数

開発時に使用する主な環境変数：

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `HOST` | `localhost` | サーバーホスト |
| `PORT` | `8080` | サーバーポート |
| `LOG_LEVEL` | `debug` | ログレベル |
| `DEV_MODE` | `true` | 開発モード |

詳細は `.env.example` を参照してください。

## 📝 VS Code開発

### 推奨拡張機能

- Go (Google)
- Go Outliner
- REST Client
- Docker
- GitLens

### デバッグ

1. VS Codeでプロジェクトを開く
2. `F5`でデバッガーを起動
3. ブレークポイントを設定してデバッグ

### タスク実行

- `Ctrl+Shift+P` → `Tasks: Run Task` でMakeタスクを実行

## 🔄 開発ワークフロー

### 一般的な開発の流れ

1. **機能ブランチを作成**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **開発環境を起動**
   ```bash
   make dev
   ```

3. **コードを実装**
   - ホットリロードにより変更が自動反映

4. **テストを作成・実行**
   ```bash
   make test
   ```

5. **コード品質チェック**
   ```bash
   make audit
   ```

6. **コミット・プッシュ**
   ```bash
   git add .
   git commit -m "Add new feature"
   git push origin feature/new-feature
   ```

### ホットリロード

開発サーバー起動時、以下のファイル変更を監視して自動再起動：

- `*.go` ファイル
- テンプレートファイル（将来対応）

## 🐛 トラブルシューティング

### よくある問題

**Q: `make dev` でサーバーが起動しない**
A: `.env`ファイルが存在することを確認。`./scripts/setup.sh`を実行して環境をセットアップ。

**Q: ホットリロードが動作しない**
A: Airがインストールされているか確認。`go install github.com/air-verse/air@latest`でインストール。

**Q: テストが失敗する**
A: 依存関係が最新か確認。`go mod tidy`を実行。

**Q: ポート8080が使用中**
A: `.env`ファイルで`PORT`を別の値に変更。

### ログ確認

開発サーバーのログは標準出力に表示されます。ログレベルは環境変数`LOG_LEVEL`で調整可能：

- `debug`: 詳細情報
- `info`: 一般情報
- `warn`: 警告
- `error`: エラーのみ

## 📞 サポート

問題が発生した場合：

1. このドキュメントを確認
2. `make help`でコマンド一覧を確認
3. GitHubのIssuesで報告