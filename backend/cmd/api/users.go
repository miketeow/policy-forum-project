package main

import (
	"log"
	"net/http"
	"policy-forum-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (app *application) getUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	posts, err := app.db.ListPostsByUser(r.Context(), store.ListPostsByUserParams{
		UserID:        userID,
		CurrentUserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:         int32(pagination.Limit),
		Cursor:        pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		log.Printf("Failed to get user posts: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if posts == nil {
		posts = []store.ListPostsByUserRow{}
	}

	writeJSON(w, http.StatusOK, posts)
}

func (app *application) getUserUpvotedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	posts, err := app.db.ListUpvotedPostsByUser(r.Context(), store.ListUpvotedPostsByUserParams{
		UserID: userID,
		Limit:  int32(pagination.Limit),
		Cursor: pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		log.Printf("Failed to get upvoted posts: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if posts == nil {
		posts = []store.ListUpvotedPostsByUserRow{}
	}

	writeJSON(w, http.StatusOK, posts)
}

func (app *application) getUserCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	posts, err := app.db.ListCommentsByUser(r.Context(), store.ListCommentsByUserParams{
		UserID:        userID,
		CurrentUserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:         int32(pagination.Limit),
		Cursor:        pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		log.Printf("Failed to get user comments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if posts == nil {
		posts = []store.ListCommentsByUserRow{}
	}

	writeJSON(w, http.StatusOK, posts)
}

func (app *application) getUserUpvotedCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	posts, err := app.db.ListUpvotedCommentsByUser(r.Context(), store.ListUpvotedCommentsByUserParams{
		UserID: userID,
		Limit:  int32(pagination.Limit),
		Cursor: pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		log.Printf("Failed to get upvoted comments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if posts == nil {
		posts = []store.ListUpvotedCommentsByUserRow{}
	}

	writeJSON(w, http.StatusOK, posts)
}
