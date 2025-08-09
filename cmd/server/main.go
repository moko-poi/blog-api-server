package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/moko-poi/blog-api-server/internal/api"
	"github.com/moko-poi/blog-api-server/internal/config"
	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
)

// エラーハンドリングが行えないため、main関数はシンプルに保つ
func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdin, os.Stdout, os.Stderr, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// run関数はメイン処理を担当
// OSの基本要素（環境変数、標準入力、引数）を引数として受け取ることで、テスタビリティを向上
func run(
	ctx context.Context,
	getenv func(string) string, // 環境変数を注入することでテスト時に制御可能
	stdin io.Reader, // 標準入力を注入することでテスト時に制御可能
	stdout io.Writer, // 標準出力を注入することでテスト時に制御可能
	stderr io.Writer, // 標準エラー出力を注入することでテスト時に制御可能
	args []string, // コマンドライン引数を注入することでテスト時に制御可能
) error {
	// GracefulShutdownのためのコンテキスト設定
	// signal.NotifyContextを使用してCtrl+CやSIGTERMを受け取る
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// 設定読み込み
	cfg, err := config.Load(getenv)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ロガーの初期化 - 出力先を注入可能にすることでテスト時はログを制御可能
	log := logger.New(stdout, cfg.LogLevel)

	// ストレージの初期化 - インメモリストアを利用（本番環境では他の実装に差し替え可能）
	blogstore := store.NewMemoryBlogStore()

	// サーバーの初期化 - 必要なコンポーネントを注入
	server, err := api.NewServer(
		log,
		cfg,
		blogstore,
	)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	return server.Start(ctx)
}
