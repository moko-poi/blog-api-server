package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
)

func TestAddRoutes(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	blogStore := store.NewMemoryBlogStore()
	mux := http.NewServeMux()

	addRoutes(mux, log, blogStore)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "healthz endpoint",
			method:         http.MethodGet,
			path:           "/healthz",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "readyz endpoint",
			method:         http.MethodGet,
			path:           "/readyz",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET blogs endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST blogs endpoint",
			method:         http.MethodPost,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusBadRequest, // Will fail validation with empty body
		},
		{
			name:           "unsupported method on blogs endpoint",
			method:         http.MethodPatch,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET specific blog endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/blogs/non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT specific blog endpoint",
			method:         http.MethodPut,
			path:           "/api/v1/blogs/non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "DELETE specific blog endpoint",
			method:         http.MethodDelete,
			path:           "/api/v1/blogs/non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAddRoutes_BlogsEndpointMethodRouting(t *testing.T) {
	log := logger.New(io.Discard, slog.LevelError)
	blogStore := store.NewMemoryBlogStore()
	mux := http.NewServeMux()

	addRoutes(mux, log, blogStore)

	// Test that the routing logic correctly delegates to the right handlers
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "GET blogs",
			method:         http.MethodGet,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusOK,
			description:    "Should route to handleBlogsGet",
		},
		{
			name:           "POST blogs",
			method:         http.MethodPost,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusBadRequest,
			description:    "Should route to handleBlogsCreate (fails validation with empty body)",
		},
		{
			name:           "PUT blogs",
			method:         http.MethodPut,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should return method not allowed",
		},
		{
			name:           "DELETE blogs",
			method:         http.MethodDelete,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should return method not allowed",
		},
		{
			name:           "HEAD blogs",
			method:         http.MethodHead,
			path:           "/api/v1/blogs",
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should return method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.description, tt.expectedStatus, w.Code)
			}
		})
	}
}