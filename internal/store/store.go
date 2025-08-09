package store

import (
	"context"
	"errors"
	"sync"

	"github.com/moko-poi/blog-api-server/pkg/models"
)

var (
	// ErrNotFound is returned when a blog is not found
	ErrNotFound = errors.New("blog not found")
)

// BlogStore defines the interface for blog storage operations
// Following Mat Ryer's pattern of simple, focused interfaces
type BlogStore interface {
	Create(ctx context.Context, blog *models.Blog) error
	GetByID(ctx context.Context, id string) (*models.Blog, error)
	GetAll(ctx context.Context) ([]*models.Blog, error)
	GetByAuthor(ctx context.Context, author string) ([]*models.Blog, error)
	Update(ctx context.Context, id string, blog *models.Blog) error
	Delete(ctx context.Context, id string) error
}

// MemoryBlogStore is an in-memory implementation of BlogStore
// Suitable for development and testing, but not for production
type MemoryBlogStore struct {
	mu    sync.RWMutex
	blogs map[string]*models.Blog
}

// NewMemoryBlogStore creates a new in-memory blog store
func NewMemoryBlogStore() *MemoryBlogStore {
	return &MemoryBlogStore{
		blogs: make(map[string]*models.Blog),
	}
}

// Create stores a new blog
func (s *MemoryBlogStore) Create(ctx context.Context, blog *models.Blog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blogs[blog.ID] = blog
	return nil
}

// GetByID retrieves a blog by its ID
func (s *MemoryBlogStore) GetByID(ctx context.Context, id string) (*models.Blog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blog, exists := s.blogs[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent modification
	blogCopy := *blog
	return &blogCopy, nil
}

// GetAll retrieves all blogs
func (s *MemoryBlogStore) GetAll(ctx context.Context) ([]*models.Blog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blogs := make([]*models.Blog, 0, len(s.blogs))
	for _, blog := range s.blogs {
		// Return copies to prevent modification
		blogCopy := *blog
		blogs = append(blogs, &blogCopy)
	}

	return blogs, nil
}

// GetByAuthor retrieves all blogs by a specific author
func (s *MemoryBlogStore) GetByAuthor(ctx context.Context, author string) ([]*models.Blog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var blogs []*models.Blog
	for _, blog := range s.blogs {
		if blog.Author == author {
			// Return a copy to prevent modification
			blogCopy := *blog
			blogs = append(blogs, &blogCopy)
		}
	}

	return blogs, nil
}

// Update updates an existing blog
func (s *MemoryBlogStore) Update(ctx context.Context, id string, blog *models.Blog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.blogs[id]; !exists {
		return ErrNotFound
	}

	s.blogs[id] = blog
	return nil
}

// Delete removes a blog by its ID
func (s *MemoryBlogStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.blogs[id]; !exists {
		return ErrNotFound
	}

	delete(s.blogs, id)
	return nil
}
