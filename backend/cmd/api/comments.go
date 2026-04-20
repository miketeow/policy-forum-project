package main

import (
	"encoding/json"
	"log"
	"net/http"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCommentRequest struct {
	Content  string     `json:"content"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Content == "" {
		writeJSONError(w, http.StatusBadRequest, "Comment cannot be empty")
		return
	}

	var nullParentID pgtype.UUID
	if req.ParentID != nil {
		nullParentID = pgtype.UUID{
			Bytes: *req.ParentID,
			Valid: true,
		}
	} else {
		nullParentID = pgtype.UUID{
			Valid: false,
		}
	}

	now := time.Now().UTC()

	args := store.CreateCommentsParams{
		ID:        uuid.New(),
		PostID:    postId,
		UserID:    userID,
		ParentID:  nullParentID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	comment, err := app.db.CreateComments(r.Context(), args)
	if err != nil {
		log.Printf("Failed to create comment")
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (app *application) getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	comments, err := app.db.GetCommentsByPostID(r.Context(), postId)
	if err != nil {
		log.Printf("Failed to fetch comments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// preventing null to the frontend, return empty array instead
	if comments == nil {
		comments = []store.GetCommentsByPostIDRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}
