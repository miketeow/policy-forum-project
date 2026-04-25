package main

import (
	"encoding/json"
	"log"
	"net/http"
	"policy-forum-backend/internal/store"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreatePostReqeust struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

type UpdatePostReqeust struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

type PaginationRequest struct {
	Limit  int
	Cursor time.Time
	Sort   string
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	// get user id from the jwt
	userId, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// parse the client payload
	var req CreatePostReqeust
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	category := req.Category

	if category == "" {
		category = "PENDING"
	}

	now := time.Now().UTC()
	// prepare the database arguments
	args := store.CreatePostParams{
		ID:        uuid.New(),
		UserID:    userId,
		Title:     req.Title,
		Content:   req.Content,
		Category:  store.PostCategory(category),
		CreatedAt: now,
		UpdatedAt: now,
	}
	// execuet the insert
	post, err := app.db.CreatePost(r.Context(), args)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)

}

func (app *application) listPostHandler(w http.ResponseWriter, r *http.Request) {
	var posts []store.ListPostsByNewestRow
	var err error
	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	if sortOrder == "asc" {
		oldestPost, dbErr := app.db.ListPostsByOldest(r.Context(), store.ListPostsByOldestParams{
			Limit: int32(pagination.Limit),
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
		})
		err = dbErr

		for _, p := range oldestPost {
			posts = append(posts, store.ListPostsByNewestRow(p))
		}
	} else {

		posts, err = app.db.ListPostsByNewest(r.Context(), store.ListPostsByNewestParams{
			Limit: int32(pagination.Limit),
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
		})
	}
	if err != nil {
		log.Printf("DB Error in listPostHandler: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Failed to list posts")
		return
	}
	// if no posts exist, make sure to return empty array instead of null
	if posts == nil {
		posts = []store.ListPostsByNewestRow{}
	}
	writeJSON(w, http.StatusOK, posts)
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	// extract id from params
	idParam := r.PathValue("postId")

	postId, err := uuid.Parse(idParam)

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	// fetch from database
	post, err := app.db.GetPostByID(r.Context(), postId)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Post not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdatePostReqeust
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	updatedPost, err := app.db.UpdatePost(r.Context(), store.UpdatePostParams{
		ID:        postId,
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Category:  store.PostCategory(req.Category),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		log.Printf("Failed to update post: %v", err)
		writeJSONError(w, http.StatusForbidden, "Not authorized to edit this post, or post does not exist")
		return
	}

	writeJSON(w, http.StatusOK, updatedPost)

}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = app.db.DeletePost(r.Context(), store.DeletePostParams{
		ID:     postId,
		UserID: userID,
	})

	if err != nil {
		log.Printf("Failed to delete post: %v", err)
		writeJSONError(w, http.StatusForbidden, "Not authorized to delete this post, or post does not exist")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parsePagination(r *http.Request) PaginationRequest {
	// default
	req := PaginationRequest{
		Limit: 20,
		Sort:  "desc",
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			req.Limit = l
		}
	}

	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05.999999Z07:00",
			"2006-01-02T15:04:05.999999", // Common Postgres format (no Z)
			"2006-01-02 15:04:05.999999", // Space instead of T
		}

		var parsed time.Time
		var err error
		for _, layout := range layouts {
			parsed, err = time.Parse(layout, cursorStr)
			if err == nil {
				req.Cursor = parsed.UTC()
				break
			}
		}

		if err != nil {
			log.Printf("[PAGINATION ERROR] Failed to parse cursor '%s'. Error: %v", cursorStr, err)
		}
	}

	return req
}
