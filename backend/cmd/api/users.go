package main

import (
	"errors"
	"fmt"
	"net/http"
	"policy-forum-backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (app *application) getUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	hasCursor := !pagination.Cursor.IsZero()

	posts, err := app.db.ListPostsByUser(r.Context(), store.ListPostsByUserParams{
		UserID:        userID,
		CurrentUserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:         int32(pagination.Limit),
		Cursor:        pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch user's posts from database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if posts == nil {
		posts = []store.ListPostsByUserRow{}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"posts": posts})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserUpvotedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	hasCursor := !pagination.Cursor.IsZero()

	upvotedPosts, err := app.db.ListUpvotedPostsByUser(r.Context(), store.ListUpvotedPostsByUserParams{
		UserID: userID,
		Limit:  int32(pagination.Limit),
		Cursor: pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch user's upvoted posts from database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if upvotedPosts == nil {
		upvotedPosts = []store.ListUpvotedPostsByUserRow{}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"posts": upvotedPosts})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	hasCursor := !pagination.Cursor.IsZero()

	comments, err := app.db.ListCommentsByUser(r.Context(), store.ListCommentsByUserParams{
		UserID:        userID,
		CurrentUserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:         int32(pagination.Limit),
		Cursor:        pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch user's comments from database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if comments == nil {
		comments = []store.ListCommentsByUserRow{}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserUpvotedCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	hasCursor := !pagination.Cursor.IsZero()

	upvotedComments, err := app.db.ListUpvotedCommentsByUser(r.Context(), store.ListUpvotedCommentsByUserParams{
		UserID: userID,
		Limit:  int32(pagination.Limit),
		Cursor: pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch user's upvoted comments from database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if upvotedComments == nil {
		upvotedComments = []store.ListUpvotedCommentsByUserRow{}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": upvotedComments})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from contex")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// fetch user profile from database
	user, err := app.db.GetUserByID(r.Context(), userID)
	if err != nil {
		// differentiate between not found and database error
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}
		wrappedErr := fmt.Errorf("failed to fetch user profile: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
