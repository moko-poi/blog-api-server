package api

import (
	"net/http"
	"time"

	"github.com/moko-poi/blog-api-server/internal/logger"
)

// loggingMiddleware logs HTTP requests
// Mat Ryerのアダプターパターン: ミドルウェアは依存関係を受け取り、
// http.Handler -> http.Handler の関数を返す
// これにより、ミドルウェアで必要な依存関係（ここではlogger）を注入可能
func loggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// レスポンスライターをラップしてステータスコードをキャプチャ
			// Mat Ryerのパターン: 構造化ログでリクエスト詳細を記録
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // デフォルトステータス
			}

			// 次のハンドラーを実行
			next.ServeHTTP(wrapped, r)

			// リクエスト処理時間を測定
			duration := time.Since(start)

			// 構造化ログでリクエスト情報を記録
			// キー・バリュー形式で後の解析が容易
			log.Info(r.Context(), "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
// http.ResponseWriterはステータスコードを取得する方法がないため、
// ラッパーを作成してWriteHeader呼び出し時にキャプチャ
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// corsMiddleware adds CORS headers
// CORS（Cross-Origin Resource Sharing）対応
// フロントエンドアプリケーションからのAPIアクセスを可能にする
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 本番環境では "*" ではなく、特定のオリジンを指定することを推奨
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// プリフライトリクエスト（OPTIONS）への対応
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// panicRecoveryMiddleware recovers from panics and returns a 500 error
// Mat Ryerのパターン: パニック発生時の適切な処理
// サーバークラッシュを防ぎ、ログに記録して適切なエラーレスポンスを返す
func panicRecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// defer でパニックをキャッチ
			defer func() {
				if err := recover(); err != nil {
					// パニック詳細をログに記録
					log.Error(r.Context(), "panic recovered",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
					)

					// クライアントには内部エラーとして500を返す
					// セキュリティ上、パニックの詳細は隠蔽
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					response := ErrorResponse{
						Error: "Internal server error",
					}
					encode(w, r, http.StatusInternalServerError, response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ratelimitMiddleware is a simple in-memory rate limiter
// レート制限機能 - DoS攻撃対策
// Mat Ryerの注記: 本番環境ではRedisなど外部ストアを使用すべき
func ratelimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// シンプルなレート制限ロジックをここに実装
			// 現在はパススルーだが、本番環境では以下のような実装が必要:
			// - IPアドレス単位での制限
			// - トークンバケットアルゴリズム
			// - Redis等を使った分散対応
			next.ServeHTTP(w, r)
		})
	}
}
