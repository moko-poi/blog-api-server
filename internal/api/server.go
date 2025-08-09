package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/moko-poi/blog-api-server/internal/config"
	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
)

// ServerはAPIサーバーの構造体
// 必要なコンポーネント（ロガー、設定、ストア）を注入して初期化する
type Server struct {
	config    *config.Config
	logger    *logger.Logger
	blogStore store.BlogStore
	server    *http.Server
}

// コストラクタでは全ての依存関係を引数として受け取る
// これにより依存関係が明確にな、テスト時に必要な依存関係だけを渡すことができる
func NewServer(
	log *logger.Logger,
	cfg *config.Config,
	blogstore store.BlogStore,
) (*Server, error) {
	// http.NewServeMuxを使用してルーティングを設定
	mux := http.NewServeMux()

	// routes.goでルート定義を一箇所に集約
	// API全体の構造が一目でわかる
	addRoutes(mux, log, blogstore)

	// ミドルウェアの設定（逆順で実行される）
	// adapter patternを使用してミドをルウェア構成
	var handler http.Handler = mux
	handler = corsMiddleware()(handler)             // CORS対応
	handler = ratelimitMiddleware()(handler)        // レート制限
	handler = panicRecoveryMiddleware(log)(handler) // パニックリカバリー
	handler = loggingMiddleware(log)(handler)       // ログ出力

	// HTTPサーバーの設定
	// タイムアウト設定
	httpServer := &http.Server{
		Addr:         cfg.Address(),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,  // 読み取りタイムアウト
		WriteTimeout: cfg.WriteTimeout, // 書き込みタイムアウト
		IdleTimeout:  30 * time.Second, // アイドルタイムアウト
	}

	return &Server{
		config:    cfg,
		logger:    log,
		blogStore: blogstore,
		server:    httpServer,
	}, nil
}

// コンテキストを受け取って、Graceful shutdownに対応
func (s *Server) Start(ctx context.Context) error {
	// サーバーエラーを受信するためのチャネル
	serverErr := make(chan error, 1)

	// サーバーをgoroutineで起動
	go func() {
		s.logger.Info(ctx, "starting server", "address", s.server.Addr)

		// net.Listen を明示的に呼び出すことで、ポート番号が0の場合の対応などが可能
		listener, err := net.Listen("tcp", s.server.Addr)
		if err != nil {
			serverErr <- fmt.Errorf("failed to create listener: %w", err)
			return
		}

		// http.ErrServerClosedはサーバーが正常にシャットダウン時のエラーなので除外
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	// サーバーエラーまたはコンテキストキャンセルを待機
	// select文でシグナル待ちとエラー処理同時に行う
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		s.logger.Info(ctx, "shutdown signal received")
		return s.shutdown() // コンテキストを渡してシャットダウン
	}
}

// グレースフルシャットダウンの実装
// 進行中のリクエストを完了させてからサーバーを停止
func (s *Server) shutdown() error {
	// シャットダウン用のタイムアウト付きコンテキストを作成
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	s.logger.Info(shutdownCtx, "shutting down server", "timeout", s.config.ShutdownTimeout)

	// Shutdownメソッドは進行中のリクエストを完了するまで待機する
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info(shutdownCtx, "server shutdown complete")
	return nil
}

// サーバーの準備完了を待つヘルパー関数
// テスト時にサーバーが起動するまで待機するために使用
func waitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	startTime := time.Now()

	// ポーリングによる準備完了チェック
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil // 準備完了
		}
		if resp != nil {
			resp.Body.Close()
		}

		// コンテキストキャンセルまたはタイムアウトをチェック
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond) // 短い間隔でリトライ
		}
	}
}
