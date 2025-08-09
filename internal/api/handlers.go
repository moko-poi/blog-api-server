package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/moko-poi/blog-api-server/internal/logger"
	"github.com/moko-poi/blog-api-server/internal/store"
	"github.com/moko-poi/blog-api-server/internal/domain"
)

// handleHealthz returns a simple health check
func handleHealthz(log *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "ok",
		}
		if err := encode(w, r, http.StatusOK, response); err != nil {
			log.Error(r.Context(), "failed to encode health response", "error", err)
		}
	})
}

// handleBlogsCreate creates a new blog post
func handleBlogsCreate(log *logger.Logger, blogStore store.BlogStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		req, problems, err := decodeValid[domain.CreateBlogRequest](r)
		if err != nil {
			if problems != nil {
				response := ErrorResponse{
					Error:    "Validation failed",
					Problems: problems,
				}
				encode(w, r, http.StatusBadRequest, response)
				return
			}
			log.Error(r.Context(), "failed to decode request", "error", err)
			response := ErrorResponse{Error: "Invalid request body"}
			encode(w, r, http.StatusBadRequest, response)
			return
		}

		blog := domain.NewBlog(req)
		if err := blogStore.Create(r.Context(), blog); err != nil {
			log.Error(r.Context(), "failed to create blog", "error", err)
			response := ErrorResponse{Error: "Failed to create blog"}
			encode(w, r, http.StatusInternalServerError, response)
			return
		}

		log.Info(r.Context(), "blog created", "id", blog.ID, "title", blog.Title)
		encode(w, r, http.StatusCreated, blog)
	})
}

// handleBlogsGet retrieves all blogs or filters by author
func handleBlogsGet(log *logger.Logger, blogStore store.BlogStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		author := r.URL.Query().Get("author")

		var blogs []*domain.Blog
		var err error

		if author != "" {
			blogs, err = blogStore.GetByAuthor(r.Context(), author)
		} else {
			blogs, err = blogStore.GetAll(r.Context())
		}

		if err != nil {
			log.Error(r.Context(), "failed to get blogs", "error", err)
			response := ErrorResponse{Error: "Failed to retrieve blogs"}
			encode(w, r, http.StatusInternalServerError, response)
			return
		}

		encode(w, r, http.StatusOK, blogs)
	})
}

// handleBlogsByID handles operations on a specific blog (GET, PUT, DELETE)
func handleBlogsByID(log *logger.Logger, blogStore store.BlogStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/blogs/")
		if path == "" || strings.Contains(path, "/") {
			response := ErrorResponse{Error: "Invalid blog ID"}
			encode(w, r, http.StatusBadRequest, response)
			return
		}
		id := path

		switch r.Method {
		case http.MethodGet:
			handleBlogGet(log, blogStore, id, w, r)
		case http.MethodPut:
			handleBlogUpdate(log, blogStore, id, w, r)
		case http.MethodDelete:
			handleBlogDelete(log, blogStore, id, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleBlogGet(log *logger.Logger, blogStore store.BlogStore, id string, w http.ResponseWriter, r *http.Request) {
	blog, err := blogStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response := ErrorResponse{Error: "Blog not found"}
			encode(w, r, http.StatusNotFound, response)
			return
		}
		log.Error(r.Context(), "failed to get blog", "error", err, "id", id)
		response := ErrorResponse{Error: "Failed to retrieve blog"}
		encode(w, r, http.StatusInternalServerError, response)
		return
	}

	encode(w, r, http.StatusOK, blog)
}

func handleBlogUpdate(log *logger.Logger, blogStore store.BlogStore, id string, w http.ResponseWriter, r *http.Request) {
	// First check if blog exists
	existingBlog, err := blogStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response := ErrorResponse{Error: "Blog not found"}
			encode(w, r, http.StatusNotFound, response)
			return
		}
		log.Error(r.Context(), "failed to get blog for update", "error", err, "id", id)
		response := ErrorResponse{Error: "Failed to retrieve blog"}
		encode(w, r, http.StatusInternalServerError, response)
		return
	}

	req, problems, err := decodeValid[domain.UpdateBlogRequest](r)
	if err != nil {
		if problems != nil {
			response := ErrorResponse{
				Error:    "Validation failed",
				Problems: problems,
			}
			encode(w, r, http.StatusBadRequest, response)
			return
		}
		log.Error(r.Context(), "failed to decode update request", "error", err)
		response := ErrorResponse{Error: "Invalid request body"}
		encode(w, r, http.StatusBadRequest, response)
		return
	}

	// Update the blog
	existingBlog.Update(req)
	if err := blogStore.Update(r.Context(), id, existingBlog); err != nil {
		log.Error(r.Context(), "failed to update blog", "error", err, "id", id)
		response := ErrorResponse{Error: "Failed to update blog"}
		encode(w, r, http.StatusInternalServerError, response)
		return
	}

	log.Info(r.Context(), "blog updated", "id", id)
	encode(w, r, http.StatusOK, existingBlog)
}

func handleBlogDelete(log *logger.Logger, blogStore store.BlogStore, id string, w http.ResponseWriter, r *http.Request) {
	if err := blogStore.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response := ErrorResponse{Error: "Blog not found"}
			encode(w, r, http.StatusNotFound, response)
			return
		}
		log.Error(r.Context(), "failed to delete blog", "error", err, "id", id)
		response := ErrorResponse{Error: "Failed to delete blog"}
		encode(w, r, http.StatusInternalServerError, response)
		return
	}

	log.Info(r.Context(), "blog deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}
