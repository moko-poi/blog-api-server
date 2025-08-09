package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moko-poi/blog-api-server/internal/logger"
)

func TestLoggingMiddleware(t *testing.T) {
	var logOutput bytes.Buffer
	log := logger.New(&logOutput, slog.LevelInfo)

	middleware := loggingMiddleware(log)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})
	
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	logContent := logOutput.String()
	if !strings.Contains(logContent, "request completed") {
		t.Error("expected log to contain 'request completed'")
	}
	if !strings.Contains(logContent, "POST") {
		t.Error("expected log to contain method 'POST'")
	}
	if !strings.Contains(logContent, "/test") {
		t.Error("expected log to contain path '/test'")
	}
	if !strings.Contains(logContent, "201") {
		t.Error("expected log to contain status code '201'")
	}
	if !strings.Contains(logContent, "test-agent") {
		t.Error("expected log to contain user agent 'test-agent'")
	}
}

func TestLoggingMiddleware_DefaultStatus(t *testing.T) {
	var logOutput bytes.Buffer
	log := logger.New(&logOutput, slog.LevelInfo)

	middleware := loggingMiddleware(log)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't explicitly set status code, should default to 200
		w.Write([]byte("test response"))
	})
	
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	logContent := logOutput.String()
	if !strings.Contains(logContent, "200") {
		t.Error("expected log to contain default status code '200'")
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	wrapper := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	wrapper.WriteHeader(http.StatusNotFound)

	if wrapper.statusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, wrapper.statusCode)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("expected underlying writer status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCorsMiddleware(t *testing.T) {
	middleware := corsMiddleware()
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	wrappedHandler := middleware(handler)

	t.Run("normal request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("expected Access-Control-Allow-Origin header to be '*'")
		}
		if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Error("expected Access-Control-Allow-Methods header")
		}
		if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
			t.Error("expected Access-Control-Allow-Headers header")
		}
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for OPTIONS request, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("expected CORS headers for OPTIONS request")
		}
	})
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	var logOutput bytes.Buffer
	log := logger.New(&logOutput, slog.LevelError)

	middleware := panicRecoveryMiddleware(log)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d after panic, got %d", http.StatusInternalServerError, w.Code)
	}

	logContent := logOutput.String()
	if !strings.Contains(logContent, "panic recovered") {
		t.Error("expected log to contain 'panic recovered'")
	}
	if !strings.Contains(logContent, "test panic") {
		t.Error("expected log to contain panic message 'test panic'")
	}

	// Check response content
	if !strings.Contains(w.Body.String(), "Internal server error") {
		t.Error("expected response to contain error message")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestPanicRecoveryMiddleware_NoPanic(t *testing.T) {
	var logOutput bytes.Buffer
	log := logger.New(&logOutput, slog.LevelError)

	middleware := panicRecoveryMiddleware(log)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("normal response"))
	})
	
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "normal response" {
		t.Errorf("expected normal response, got %q", w.Body.String())
	}

	// Should not have panic logs
	logContent := logOutput.String()
	if strings.Contains(logContent, "panic recovered") {
		t.Error("expected no panic recovery logs")
	}
}

func TestRatelimitMiddleware(t *testing.T) {
	middleware := ratelimitMiddleware()
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Currently rate limiting is a pass-through, so should work normally
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("expected success response, got %q", w.Body.String())
	}
}