package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moko-poi/blog-api-server/internal/domain"
)

func TestMemoryBlogStore_Create(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	blog := &domain.Blog{
		ID:        "test-id",
		Title:     "Test Title",
		Content:   "Test Content",
		Author:    "Test Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := store.Create(ctx, blog)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify blog was stored
	stored, err := store.GetByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error retrieving blog, got %v", err)
	}

	if stored.ID != blog.ID {
		t.Errorf("expected ID %q, got %q", blog.ID, stored.ID)
	}
	if stored.Title != blog.Title {
		t.Errorf("expected Title %q, got %q", blog.Title, stored.Title)
	}
}

func TestMemoryBlogStore_GetByID(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	blog := &domain.Blog{
		ID:        "test-id",
		Title:     "Test Title",
		Content:   "Test Content",
		Author:    "Test Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Test getting non-existent blog
	_, err := store.GetByID(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Store blog and retrieve it
	store.Create(ctx, blog)
	retrieved, err := store.GetByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's a copy (different pointer)
	if retrieved == blog {
		t.Error("expected different pointer (copy), got same pointer")
	}

	// But same content
	if retrieved.ID != blog.ID {
		t.Errorf("expected ID %q, got %q", blog.ID, retrieved.ID)
	}
	if retrieved.Title != blog.Title {
		t.Errorf("expected Title %q, got %q", blog.Title, retrieved.Title)
	}

	// Verify modifying returned blog doesn't affect stored blog
	retrieved.Title = "Modified Title"
	stored, _ := store.GetByID(ctx, "test-id")
	if stored.Title == "Modified Title" {
		t.Error("modifying returned blog affected stored blog")
	}
}

func TestMemoryBlogStore_GetAll(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	// Test empty store
	blogs, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(blogs) != 0 {
		t.Errorf("expected 0 blogs, got %d", len(blogs))
	}

	// Add multiple blogs
	blog1 := &domain.Blog{
		ID:        "id1",
		Title:     "Title 1",
		Content:   "Content 1",
		Author:    "Author 1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blog2 := &domain.Blog{
		ID:        "id2",
		Title:     "Title 2",
		Content:   "Content 2",
		Author:    "Author 2",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	store.Create(ctx, blog1)
	store.Create(ctx, blog2)

	blogs, err = store.GetAll(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(blogs) != 2 {
		t.Errorf("expected 2 blogs, got %d", len(blogs))
	}

	// Verify they're copies
	for _, blog := range blogs {
		if blog == blog1 || blog == blog2 {
			t.Error("expected different pointers (copies), got same pointer")
		}
	}

	// Verify modifying returned blogs doesn't affect stored blogs
	blogs[0].Title = "Modified Title"
	stored, _ := store.GetByID(ctx, blogs[0].ID)
	if stored.Title == "Modified Title" {
		t.Error("modifying returned blog affected stored blog")
	}
}

func TestMemoryBlogStore_GetByAuthor(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	// Test with no blogs
	blogs, err := store.GetByAuthor(ctx, "NonExistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(blogs) != 0 {
		t.Errorf("expected 0 blogs, got %d", len(blogs))
	}

	// Add blogs with different authors
	blog1 := &domain.Blog{
		ID:        "id1",
		Title:     "Title 1",
		Content:   "Content 1",
		Author:    "Author A",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blog2 := &domain.Blog{
		ID:        "id2",
		Title:     "Title 2",
		Content:   "Content 2",
		Author:    "Author B",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	blog3 := &domain.Blog{
		ID:        "id3",
		Title:     "Title 3",
		Content:   "Content 3",
		Author:    "Author A", // Same author as blog1
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	store.Create(ctx, blog1)
	store.Create(ctx, blog2)
	store.Create(ctx, blog3)

	// Get blogs by Author A
	blogs, err = store.GetByAuthor(ctx, "Author A")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(blogs) != 2 {
		t.Errorf("expected 2 blogs, got %d", len(blogs))
	}

	// Verify correct blogs were returned
	authorA := 0
	for _, blog := range blogs {
		if blog.Author == "Author A" {
			authorA++
		}
	}
	if authorA != 2 {
		t.Errorf("expected 2 blogs by Author A, got %d", authorA)
	}

	// Get blogs by Author B
	blogs, err = store.GetByAuthor(ctx, "Author B")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(blogs) != 1 {
		t.Errorf("expected 1 blog, got %d", len(blogs))
	}
	if blogs[0].Author != "Author B" {
		t.Errorf("expected Author B, got %q", blogs[0].Author)
	}
}

func TestMemoryBlogStore_Update(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	// Test updating non-existent blog
	blog := &domain.Blog{
		ID:        "non-existent",
		Title:     "Title",
		Content:   "Content",
		Author:    "Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := store.Update(ctx, "non-existent", blog)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Create and update blog
	originalBlog := &domain.Blog{
		ID:        "test-id",
		Title:     "Original Title",
		Content:   "Original Content",
		Author:    "Original Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	store.Create(ctx, originalBlog)

	updatedBlog := &domain.Blog{
		ID:        "test-id",
		Title:     "Updated Title",
		Content:   "Updated Content",
		Author:    "Original Author",
		CreatedAt: originalBlog.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}

	err = store.Update(ctx, "test-id", updatedBlog)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify update
	retrieved, err := store.GetByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error retrieving updated blog, got %v", err)
	}
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected updated title, got %q", retrieved.Title)
	}
	if retrieved.Content != "Updated Content" {
		t.Errorf("expected updated content, got %q", retrieved.Content)
	}
}

func TestMemoryBlogStore_Delete(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	// Test deleting non-existent blog
	err := store.Delete(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Create and delete blog
	blog := &domain.Blog{
		ID:        "test-id",
		Title:     "Test Title",
		Content:   "Test Content",
		Author:    "Test Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	store.Create(ctx, blog)

	// Verify blog exists
	_, err = store.GetByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected blog to exist before deletion, got %v", err)
	}

	// Delete blog
	err = store.Delete(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error deleting blog, got %v", err)
	}

	// Verify blog was deleted
	_, err = store.GetByID(ctx, "test-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestMemoryBlogStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryBlogStore()
	ctx := context.Background()

	// Test concurrent reads and writes
	blog := &domain.Blog{
		ID:        "test-id",
		Title:     "Test Title",
		Content:   "Test Content",
		Author:    "Test Author",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	done := make(chan bool, 2)

	// Goroutine 1: Write operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			testBlog := *blog
			testBlog.ID = "test-id-" + string(rune('0'+i%10))
			store.Create(ctx, &testBlog)
		}
	}()

	// Goroutine 2: Read operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			store.GetAll(ctx)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify store is still functional
	finalBlogs, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("expected no error after concurrent access, got %v", err)
	}

	// Should have some blogs from the concurrent writes
	if len(finalBlogs) == 0 {
		t.Error("expected some blogs after concurrent operations")
	}
}

func TestMemoryBlogStore_Interface(t *testing.T) {
	// Verify MemoryBlogStore implements BlogStore interface
	var _ BlogStore = (*MemoryBlogStore)(nil)
}