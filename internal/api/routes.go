package api

import (
	"net/http"

	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
)

// routes.goでAPI全体の構造を一箇所で定義
func addRoutes(
	mux *http.ServeMux,
	log *logger.Logger,
	blogStore store.BlogStore,
) {
	// ヘルスチェックエンドポイント
	mux.Handle("/healthz", handleHealthz(log))
	mux.Handle("/readyz", handleHealthz(log))

	// GET /api/v1/blogs (全ブログ取得) とPOST /api/v1/blogs (ブログ作成)
	// Go標準のmuxでは同じパスで異なるHTTPメソッドを処理するために
	// HandlerFuncで条件分岐する必要がある
	mux.HandleFunc("/api/v1/blogs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleBlogsGet(log, blogStore).ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodPost {
			handleBlogsCreate(log, blogStore).ServeHTTP(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	// GET, PUT, DELETE /api/v1/blogs/{id}
	// Go標準のmuxでは動的パスパラメータが限定的なので、プレフィックスマッチを使用
	mux.Handle("/api/v1/blogs/", handleBlogsByID(log, blogStore))
}
