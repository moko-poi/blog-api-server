package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Blog represents a blog post
// Mat Ryerのパターン: ドメインモデルは pkg/ 配下に配置
// 外部パッケージからも参照可能な公開型として定義
type Blog struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateBlogRequest represents a request to create a new blog
// Mat Ryerのパターン: リクエスト/レスポンス型をハンドラー内で定義する場合もあるが、
// 複数のハンドラーで共有する場合はmodelsパッケージに配置
type CreateBlogRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

// Valid implements the Validator interface
// Mat Ryerのシンプルバリデーションパターン
// オブジェクト自身がバリデーション責任を持ち、問題をmap[string]stringで返す
// データベースチェックなど重い処理はここでは行わず、基本的な形式チェックのみ
func (r CreateBlogRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	// タイトルのバリデーション
	if strings.TrimSpace(r.Title) == "" {
		problems["title"] = "title is required"
	}

	if len(r.Title) > 100 {
		problems["title"] = "title must be less than 100 characters"
	}

	// コンテンツのバリデーション
	if strings.TrimSpace(r.Content) == "" {
		problems["content"] = "content is required"
	}

	if len(r.Content) > 5000 {
		problems["content"] = "content must be less than 5000 characters"
	}

	// 作者のバリデーション
	if strings.TrimSpace(r.Author) == "" {
		problems["author"] = "author is required"
	}

	if len(r.Author) > 50 {
		problems["author"] = "author must be less than 50 characters"
	}

	return problems
}

// UpdateBlogRequest represents a request to update a blog
// ポインタ型を使用することで、フィールドが指定されたかどうかを判別可能
// nilの場合は更新対象外、値がある場合は更新対象として扱う
type UpdateBlogRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

// Valid implements the Validator interface
// 更新リクエストのバリデーション - 指定されたフィールドのみチェック
func (r UpdateBlogRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	// タイトルが指定されている場合のみバリデーション
	if r.Title != nil {
		if len(*r.Title) > 100 {
			problems["title"] = "title must be less than 100 characters"
		}
		if strings.TrimSpace(*r.Title) == "" {
			problems["title"] = "title cannot be empty"
		}
	}

	// コンテンツが指定されている場合のみバリデーション
	if r.Content != nil {
		if len(*r.Content) > 5000 {
			problems["content"] = "content must be less than 5000 characters"
		}
		if strings.TrimSpace(*r.Content) == "" {
			problems["content"] = "content cannot be empty"
		}
	}

	return problems
}

// NewBlog creates a new blog from a create request
// Mat Ryerのパターン: ファクトリー関数でドメインオブジェクトを生成
// IDの生成、タイムスタンプの設定、データの正規化などを一箇所で処理
func NewBlog(req CreateBlogRequest) *Blog {
	now := time.Now().UTC() // UTCで統一してタイムゾーンの問題を回避
	return &Blog{
		ID:        uuid.New().String(),            // 一意なIDを自動生成
		Title:     strings.TrimSpace(req.Title),   // 前後の空白を除去
		Content:   strings.TrimSpace(req.Content), // 前後の空白を除去
		Author:    strings.TrimSpace(req.Author),  // 前後の空白を除去
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the blog with the provided update request
// Mat Ryerのパターン: ドメインモデルがビジネスロジックを担当
// 更新処理をモデル自身のメソッドとして実装し、ビジネスルールを集約
func (b *Blog) Update(req UpdateBlogRequest) {
	// 指定されたフィールドのみ更新
	if req.Title != nil {
		b.Title = strings.TrimSpace(*req.Title)
	}
	if req.Content != nil {
		b.Content = strings.TrimSpace(*req.Content)
	}
	// 更新日時は常に現在時刻に設定
	b.UpdatedAt = time.Now().UTC()
}
