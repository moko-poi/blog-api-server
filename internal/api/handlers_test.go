package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
	"github.com/moko-poi/blog-api-server/internal/domain"
)

func TestHandleHealthz(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	handler := handleHealthz(log)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", response["status"])
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestHandleBlogsCreate(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	blogStore := store.NewMemoryBlogStore()
	handler := handleBlogsCreate(log, blogStore)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if resp.Error != "Invalid request body" {
					t.Errorf("expected error 'Invalid request body', got %q", resp.Error)
				}
			},
		},
		{
			name:   "validation error",
			method: http.MethodPost,
			body: domain.CreateBlogRequest{
				Title:   "",
				Content: "Valid content",
				Author:  "Valid author",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if resp.Error != "Validation failed" {
					t.Errorf("expected error 'Validation failed', got %q", resp.Error)
				}
				if resp.Problems == nil || resp.Problems["title"] == "" {
					t.Error("expected validation problem for title field")
				}
			},
		},
		{
			name:   "successful creation",
			method: http.MethodPost,
			body: domain.CreateBlogRequest{
				Title:   "Test Title",
				Content: "Test content",
				Author:  "Test Author",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var blog domain.Blog
				if err := json.Unmarshal(body, &blog); err != nil {
					t.Fatalf("failed to unmarshal blog response: %v", err)
				}
				if blog.ID == "" {
					t.Error("expected blog ID to be set")
				}
				if blog.Title != "Test Title" {
					t.Errorf("expected title 'Test Title', got %q", blog.Title)
				}
				if blog.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body.WriteString(str)
				} else {
					json.NewEncoder(&body).Encode(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/v1/blogs", &body)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleBlogsGet(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	blogStore := store.NewMemoryBlogStore()
	handler := handleBlogsGet(log, blogStore)

	// Add test data
	blog1 := &domain.Blog{
		ID:        "1",
		Title:     "Blog 1",
		Content:   "Content 1",
		Author:    "Author A",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blog2 := &domain.Blog{
		ID:        "2",
		Title:     "Blog 2",
		Content:   "Content 2",
		Author:    "Author B",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blog3 := &domain.Blog{
		ID:        "3",
		Title:     "Blog 3",
		Content:   "Content 3",
		Author:    "Author A",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	ctx := context.Background()
	blogStore.Create(ctx, blog1)
	blogStore.Create(ctx, blog2)
	blogStore.Create(ctx, blog3)

	tests := []struct {
		name           string
		method         string
		query          string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "wrong method",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "get all blogs",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var blogs []*domain.Blog
				if err := json.Unmarshal(body, &blogs); err != nil {
					t.Fatalf("failed to unmarshal blogs response: %v", err)
				}
				if len(blogs) != 3 {
					t.Errorf("expected 3 blogs, got %d", len(blogs))
				}
			},
		},
		{
			name:           "get blogs by author",
			method:         http.MethodGet,
			query:          "?author=Author%20A",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var blogs []*domain.Blog
				if err := json.Unmarshal(body, &blogs); err != nil {
					t.Fatalf("failed to unmarshal blogs response: %v", err)
				}
				if len(blogs) != 2 {
					t.Errorf("expected 2 blogs, got %d", len(blogs))
				}
				for _, blog := range blogs {
					if blog.Author != "Author A" {
						t.Errorf("expected author 'Author A', got %q", blog.Author)
					}
				}
			},
		},
		{
			name:           "get blogs by non-existent author",
			method:         http.MethodGet,
			query:          "?author=NonExistent",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var blogs []*domain.Blog
				if err := json.Unmarshal(body, &blogs); err != nil {
					t.Fatalf("failed to unmarshal blogs response: %v", err)
				}
				if len(blogs) != 0 {
					t.Errorf("expected 0 blogs, got %d", len(blogs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/blogs"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleBlogsByID(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	blogStore := store.NewMemoryBlogStore()
	handler := handleBlogsByID(log, blogStore)

	// Add test blog
	blog := &domain.Blog{
		ID:        "test-id",
		Title:     "Test Blog",
		Content:   "Test Content",
		Author:    "Test Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blogStore.Create(context.Background(), blog)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "invalid ID format",
			method:         http.MethodGet,
			path:           "/api/v1/blogs/",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				json.Unmarshal(body, &resp)
				if resp.Error != "Invalid blog ID" {
					t.Errorf("expected error 'Invalid blog ID', got %q", resp.Error)
				}
			},
		},
		{
			name:           "invalid ID with slash",
			method:         http.MethodGet,
			path:           "/api/v1/blogs/test/invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unsupported method",
			method:         http.MethodPatch,
			path:           "/api/v1/blogs/test-id",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "get existing blog",
			method:         http.MethodGet,
			path:           "/api/v1/blogs/test-id",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var retrievedBlog domain.Blog
				if err := json.Unmarshal(body, &retrievedBlog); err != nil {
					t.Fatalf("failed to unmarshal blog response: %v", err)
				}
				if retrievedBlog.ID != "test-id" {
					t.Errorf("expected ID 'test-id', got %q", retrievedBlog.ID)
				}
				if retrievedBlog.Title != "Test Blog" {
					t.Errorf("expected title 'Test Blog', got %q", retrievedBlog.Title)
				}
			},
		},
		{
			name:           "get non-existent blog",
			method:         http.MethodGet,
			path:           "/api/v1/blogs/non-existent",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				json.Unmarshal(body, &resp)
				if resp.Error != "Blog not found" {
					t.Errorf("expected error 'Blog not found', got %q", resp.Error)
				}
			},
		},
		{
			name:   "update existing blog",
			method: http.MethodPut,
			path:   "/api/v1/blogs/test-id",
			body: domain.UpdateBlogRequest{
				Title:   stringPtr("Updated Title"),
				Content: stringPtr("Updated Content"),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var updatedBlog domain.Blog
				if err := json.Unmarshal(body, &updatedBlog); err != nil {
					t.Fatalf("failed to unmarshal updated blog response: %v", err)
				}
				if updatedBlog.Title != "Updated Title" {
					t.Errorf("expected title 'Updated Title', got %q", updatedBlog.Title)
				}
				if updatedBlog.Content != "Updated Content" {
					t.Errorf("expected content 'Updated Content', got %q", updatedBlog.Content)
				}
			},
		},
		{
			name:   "update with validation error",
			method: http.MethodPut,
			path:   "/api/v1/blogs/test-id",
			body: domain.UpdateBlogRequest{
				Title: stringPtr(""),
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				json.Unmarshal(body, &resp)
				if resp.Error != "Validation failed" {
					t.Errorf("expected error 'Validation failed', got %q", resp.Error)
				}
				if resp.Problems["title"] == "" {
					t.Error("expected validation problem for title field")
				}
			},
		},
		{
			name:           "update non-existent blog",
			method:         http.MethodPut,
			path:           "/api/v1/blogs/non-existent",
			body:           domain.UpdateBlogRequest{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "delete existing blog",
			method:         http.MethodDelete,
			path:           "/api/v1/blogs/test-id",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent blog",
			method:         http.MethodDelete,
			path:           "/api/v1/blogs/non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset blog for each test
			if strings.Contains(tt.name, "delete existing blog") ||
				strings.Contains(tt.name, "update existing blog") ||
				strings.Contains(tt.name, "get existing blog") {
				blogStore.Create(context.Background(), blog)
			}

			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// Mock store for testing error conditions
type mockBlogStore struct {
	createError    error
	getAllError    error
	getByIDError   error
	getByAuthorError error
	updateError    error
	deleteError    error
}

func (m *mockBlogStore) Create(ctx context.Context, blog *domain.Blog) error {
	return m.createError
}

func (m *mockBlogStore) GetByID(ctx context.Context, id string) (*domain.Blog, error) {
	return nil, m.getByIDError
}

func (m *mockBlogStore) GetAll(ctx context.Context) ([]*domain.Blog, error) {
	return nil, m.getAllError
}

func (m *mockBlogStore) GetByAuthor(ctx context.Context, author string) ([]*domain.Blog, error) {
	return nil, m.getByAuthorError
}

func (m *mockBlogStore) Update(ctx context.Context, id string, blog *domain.Blog) error {
	return m.updateError
}

func (m *mockBlogStore) Delete(ctx context.Context, id string) error {
	return m.deleteError
}

func TestHandleBlogsCreate_StoreError(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	mockStore := &mockBlogStore{
		createError: errors.New("store error"),
	}
	handler := handleBlogsCreate(log, mockStore)

	reqBody := domain.CreateBlogRequest{
		Title:   "Test Title",
		Content: "Test Content",
		Author:  "Test Author",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/blogs", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error != "Failed to create blog" {
		t.Errorf("expected error 'Failed to create blog', got %q", resp.Error)
	}
}

func TestHandleBlogsGet_StoreError(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	mockStore := &mockBlogStore{
		getAllError: errors.New("store error"),
	}
	handler := handleBlogsGet(log, mockStore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/blogs", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}