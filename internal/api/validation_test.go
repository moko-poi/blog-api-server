package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moko-poi/blog-api-server/internal/domain"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		status         int
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "encode simple object",
			data:           map[string]string{"message": "hello"},
			status:         http.StatusOK,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result["message"] != "hello" {
					t.Errorf("expected message 'hello', got %q", result["message"])
				}
			},
		},
		{
			name:           "encode with different status",
			data:           ErrorResponse{Error: "test error"},
			status:         http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result.Error != "test error" {
					t.Errorf("expected error 'test error', got %q", result.Error)
				}
			},
		},
		{
			name:           "encode nil value",
			data:           nil,
			status:         http.StatusNoContent,
			expectedStatus: http.StatusNoContent,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				if strings.TrimSpace(w.Body.String()) != "null" {
					t.Errorf("expected null response for nil value, got %q", w.Body.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			err := encode(w, req, tt.status, tt.data)

			if err != nil {
				t.Fatalf("encode returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectError bool
		checkResult func(t *testing.T, result domain.CreateBlogRequest)
	}{
		{
			name:        "valid JSON",
			body:        `{"title":"Test","content":"Content","author":"Author"}`,
			expectError: false,
			checkResult: func(t *testing.T, result domain.CreateBlogRequest) {
				if result.Title != "Test" {
					t.Errorf("expected title 'Test', got %q", result.Title)
				}
				if result.Content != "Content" {
					t.Errorf("expected content 'Content', got %q", result.Content)
				}
				if result.Author != "Author" {
					t.Errorf("expected author 'Author', got %q", result.Author)
				}
			},
		},
		{
			name:        "invalid JSON",
			body:        `{"title":"Test","content":}`,
			expectError: true,
		},
		{
			name:        "empty body",
			body:        ``,
			expectError: true,
		},
		{
			name:        "wrong field types",
			body:        `{"title":123,"content":"Content","author":"Author"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			
			result, err := decode[domain.CreateBlogRequest](req)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			if !tt.expectError && tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestDecodeValid(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		expectDecodeErr bool
		expectProblems  bool
		checkProblems   func(t *testing.T, problems map[string]string)
		checkResult     func(t *testing.T, result domain.CreateBlogRequest)
	}{
		{
			name:            "valid request",
			body:            `{"title":"Test Title","content":"Test Content","author":"Test Author"}`,
			expectDecodeErr: false,
			expectProblems:  false,
			checkResult: func(t *testing.T, result domain.CreateBlogRequest) {
				if result.Title != "Test Title" {
					t.Errorf("expected title 'Test Title', got %q", result.Title)
				}
			},
		},
		{
			name:            "invalid JSON",
			body:            `{"title":"Test","content":}`,
			expectDecodeErr: true,
			expectProblems:  false,
		},
		{
			name:            "validation errors",
			body:            `{"title":"","content":"","author":""}`,
			expectDecodeErr: false,
			expectProblems:  true,
			checkProblems: func(t *testing.T, problems map[string]string) {
				if len(problems) != 3 {
					t.Errorf("expected 3 validation problems, got %d", len(problems))
				}
				if problems["title"] == "" {
					t.Error("expected title validation problem")
				}
				if problems["content"] == "" {
					t.Error("expected content validation problem")
				}
				if problems["author"] == "" {
					t.Error("expected author validation problem")
				}
			},
		},
		{
			name:            "partial validation errors",
			body:            `{"title":"Valid Title","content":"","author":"Valid Author"}`,
			expectDecodeErr: false,
			expectProblems:  true,
			checkProblems: func(t *testing.T, problems map[string]string) {
				if len(problems) != 1 {
					t.Errorf("expected 1 validation problem, got %d", len(problems))
				}
				if problems["content"] == "" {
					t.Error("expected content validation problem")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			
			result, problems, err := decodeValid[domain.CreateBlogRequest](req)

			if tt.expectDecodeErr && err == nil {
				t.Error("expected decode error but got none")
			}
			if !tt.expectDecodeErr && !tt.expectProblems && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			if tt.expectProblems && problems == nil {
				t.Error("expected validation problems but got none")
			}
			if !tt.expectProblems && problems != nil {
				t.Errorf("expected no validation problems but got: %v", problems)
			}

			if tt.checkProblems != nil && problems != nil {
				tt.checkProblems(t, problems)
			}

			if tt.checkResult != nil && !tt.expectDecodeErr && !tt.expectProblems {
				tt.checkResult(t, result)
			}
		})
	}
}

// Test that decodeValid works with UpdateBlogRequest too
func TestDecodeValid_UpdateRequest(t *testing.T) {
	body := `{"title":"Updated Title"}`
	req := httptest.NewRequest(http.MethodPut, "/test", strings.NewReader(body))
	
	result, problems, err := decodeValid[domain.UpdateBlogRequest](req)

	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	if problems != nil {
		t.Errorf("expected no validation problems but got: %v", problems)
	}

	if result.Title == nil {
		t.Error("expected title field to be set")
	} else if *result.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", *result.Title)
	}
}

func TestErrorResponse(t *testing.T) {
	// Test ErrorResponse JSON marshaling
	response := ErrorResponse{
		Error:    "Test error",
		Problems: map[string]string{"field": "problem"},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal ErrorResponse: %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if unmarshaled.Error != "Test error" {
		t.Errorf("expected error 'Test error', got %q", unmarshaled.Error)
	}

	if len(unmarshaled.Problems) != 1 {
		t.Errorf("expected 1 problem, got %d", len(unmarshaled.Problems))
	}

	if unmarshaled.Problems["field"] != "problem" {
		t.Errorf("expected problem 'problem', got %q", unmarshaled.Problems["field"])
	}
}

// Test ErrorResponse without problems field
func TestErrorResponse_NoProblems(t *testing.T) {
	response := ErrorResponse{
		Error: "Simple error",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal ErrorResponse: %v", err)
	}

	// Should not include problems field when empty
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"error":"Simple error"`) {
		t.Error("expected error field in JSON")
	}

	// With omitempty, problems should not appear
	if strings.Contains(jsonStr, "problems") {
		t.Error("expected problems field to be omitted when empty")
	}
}