package main

import (
	"encoding/json"
	"log"
	"net/http"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
)

type CreatePostReqeust struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
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
	args := store.ListPostsParams{
		Limit:  50,
		Offset: 0,
	}

	posts, err := app.db.ListPosts(r.Context(), args)
	if err != nil {
		log.Printf("Failed to list posts: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// if no posts exist, make sure to return empty array instead of null
	if posts == nil {
		posts = []store.ListPostsRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)

}
