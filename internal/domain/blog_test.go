package domain

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCreateBlogRequest_Valid(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateBlogRequest
		wantErrs []string
	}{
		{
			name: "valid request",
			req: CreateBlogRequest{
				Title:   "Valid Title",
				Content: "Valid content",
				Author:  "Valid Author",
			},
			wantErrs: nil,
		},
		{
			name: "empty title",
			req: CreateBlogRequest{
				Title:   "",
				Content: "Valid content",
				Author:  "Valid Author",
			},
			wantErrs: []string{"title"},
		},
		{
			name: "whitespace only title",
			req: CreateBlogRequest{
				Title:   "   ",
				Content: "Valid content",
				Author:  "Valid Author",
			},
			wantErrs: []string{"title"},
		},
		{
			name: "title too long",
			req: CreateBlogRequest{
				Title:   strings.Repeat("a", 101),
				Content: "Valid content",
				Author:  "Valid Author",
			},
			wantErrs: []string{"title"},
		},
		{
			name: "empty content",
			req: CreateBlogRequest{
				Title:   "Valid Title",
				Content: "",
				Author:  "Valid Author",
			},
			wantErrs: []string{"content"},
		},
		{
			name: "content too long",
			req: CreateBlogRequest{
				Title:   "Valid Title",
				Content: strings.Repeat("a", 5001),
				Author:  "Valid Author",
			},
			wantErrs: []string{"content"},
		},
		{
			name: "empty author",
			req: CreateBlogRequest{
				Title:   "Valid Title",
				Content: "Valid content",
				Author:  "",
			},
			wantErrs: []string{"author"},
		},
		{
			name: "author too long",
			req: CreateBlogRequest{
				Title:   "Valid Title",
				Content: "Valid content",
				Author:  strings.Repeat("a", 51),
			},
			wantErrs: []string{"author"},
		},
		{
			name: "multiple validation errors",
			req: CreateBlogRequest{
				Title:   "",
				Content: "",
				Author:  "",
			},
			wantErrs: []string{"title", "content", "author"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := tt.req.Valid(context.Background())
			
			if tt.wantErrs == nil && len(problems) > 0 {
				t.Errorf("expected no validation errors, got: %v", problems)
				return
			}
			
			for _, wantErr := range tt.wantErrs {
				if _, exists := problems[wantErr]; !exists {
					t.Errorf("expected validation error for field %q, but it was not found", wantErr)
				}
			}
			
			if len(problems) != len(tt.wantErrs) {
				t.Errorf("expected %d validation errors, got %d: %v", len(tt.wantErrs), len(problems), problems)
			}
		})
	}
}

func TestUpdateBlogRequest_Valid(t *testing.T) {
	validTitle := "Valid Title"
	emptyTitle := ""
	longTitle := strings.Repeat("a", 101)
	validContent := "Valid content"
	emptyContent := ""
	longContent := strings.Repeat("a", 5001)

	tests := []struct {
		name     string
		req      UpdateBlogRequest
		wantErrs []string
	}{
		{
			name:     "empty request",
			req:      UpdateBlogRequest{},
			wantErrs: nil,
		},
		{
			name: "valid title update",
			req: UpdateBlogRequest{
				Title: &validTitle,
			},
			wantErrs: nil,
		},
		{
			name: "valid content update",
			req: UpdateBlogRequest{
				Content: &validContent,
			},
			wantErrs: nil,
		},
		{
			name: "valid title and content update",
			req: UpdateBlogRequest{
				Title:   &validTitle,
				Content: &validContent,
			},
			wantErrs: nil,
		},
		{
			name: "empty title",
			req: UpdateBlogRequest{
				Title: &emptyTitle,
			},
			wantErrs: []string{"title"},
		},
		{
			name: "title too long",
			req: UpdateBlogRequest{
				Title: &longTitle,
			},
			wantErrs: []string{"title"},
		},
		{
			name: "empty content",
			req: UpdateBlogRequest{
				Content: &emptyContent,
			},
			wantErrs: []string{"content"},
		},
		{
			name: "content too long",
			req: UpdateBlogRequest{
				Content: &longContent,
			},
			wantErrs: []string{"content"},
		},
		{
			name: "multiple validation errors",
			req: UpdateBlogRequest{
				Title:   &emptyTitle,
				Content: &emptyContent,
			},
			wantErrs: []string{"title", "content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := tt.req.Valid(context.Background())
			
			if tt.wantErrs == nil && len(problems) > 0 {
				t.Errorf("expected no validation errors, got: %v", problems)
				return
			}
			
			for _, wantErr := range tt.wantErrs {
				if _, exists := problems[wantErr]; !exists {
					t.Errorf("expected validation error for field %q, but it was not found", wantErr)
				}
			}
			
			if len(problems) != len(tt.wantErrs) {
				t.Errorf("expected %d validation errors, got %d: %v", len(tt.wantErrs), len(problems), problems)
			}
		})
	}
}

func TestNewBlog(t *testing.T) {
	req := CreateBlogRequest{
		Title:   "  Test Title  ",
		Content: "  Test Content  ",
		Author:  "  Test Author  ",
	}

	blog := NewBlog(req)

	if blog.ID == "" {
		t.Error("expected blog ID to be generated, got empty string")
	}

	if blog.Title != "Test Title" {
		t.Errorf("expected title to be trimmed, got %q", blog.Title)
	}

	if blog.Content != "Test Content" {
		t.Errorf("expected content to be trimmed, got %q", blog.Content)
	}

	if blog.Author != "Test Author" {
		t.Errorf("expected author to be trimmed, got %q", blog.Author)
	}

	if blog.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if blog.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	if !blog.CreatedAt.Equal(blog.UpdatedAt) {
		t.Error("expected CreatedAt and UpdatedAt to be equal for new blog")
	}

	if blog.CreatedAt.Location() != time.UTC {
		t.Error("expected timestamps to be in UTC")
	}
}

func TestBlog_Update(t *testing.T) {
	blog := &Blog{
		ID:        "test-id",
		Title:     "Original Title",
		Content:   "Original Content",
		Author:    "Original Author",
		CreatedAt: time.Now().UTC().Add(-time.Hour),
		UpdatedAt: time.Now().UTC().Add(-time.Hour),
	}

	originalCreatedAt := blog.CreatedAt
	time.Sleep(time.Millisecond) // Ensure different timestamp

	tests := []struct {
		name           string
		req            UpdateBlogRequest
		expectedTitle  string
		expectedContent string
	}{
		{
			name:            "no updates",
			req:             UpdateBlogRequest{},
			expectedTitle:   "Original Title",
			expectedContent: "Original Content",
		},
		{
			name: "update title only",
			req: UpdateBlogRequest{
				Title: stringPtr("  New Title  "),
			},
			expectedTitle:   "New Title",
			expectedContent: "Original Content",
		},
		{
			name: "update content only",
			req: UpdateBlogRequest{
				Content: stringPtr("  New Content  "),
			},
			expectedTitle:   "Original Title",
			expectedContent: "New Content",
		},
		{
			name: "update both title and content",
			req: UpdateBlogRequest{
				Title:   stringPtr("  Updated Title  "),
				Content: stringPtr("  Updated Content  "),
			},
			expectedTitle:   "Updated Title",
			expectedContent: "Updated Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the blog for this test
			testBlog := *blog
			
			testBlog.Update(tt.req)

			if testBlog.Title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, testBlog.Title)
			}

			if testBlog.Content != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, testBlog.Content)
			}

			if testBlog.Author != "Original Author" {
				t.Errorf("expected author to remain unchanged, got %q", testBlog.Author)
			}

			if !testBlog.CreatedAt.Equal(originalCreatedAt) {
				t.Error("expected CreatedAt to remain unchanged")
			}

			if !testBlog.UpdatedAt.After(originalCreatedAt) {
				t.Error("expected UpdatedAt to be updated")
			}

			if testBlog.UpdatedAt.Location() != time.UTC {
				t.Error("expected UpdatedAt to be in UTC")
			}
		})
	}
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}