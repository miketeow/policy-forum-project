package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCommentRequest struct {
	Content  string     `json:"content" validate:"required,min=1,max=10000"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

type VoteCommentRequest struct {
	Vote int16 `json:"vote" validate:"required,oneof=1 -1"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
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
		wrappedErr := fmt.Errorf("failed to create comment in database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// declare parent slice
	var comments []store.ListCommentsByNewestRow
	var err error

	var currentUserID pgtype.UUID

	if userID, ok := r.Context().Value(userIDKey).(uuid.UUID); ok {
		currentUserID = pgtype.UUID{Bytes: userID, Valid: true}
	}

	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// pagination
	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	hasCursor := !pagination.Cursor.IsZero()

	// look for optional parent ID
	var nullParentID pgtype.UUID
	parentIDStr := r.URL.Query().Get("parentId")
	if parentIDStr != "" {
		pId, err := uuid.Parse(parentIDStr)
		if err == nil {
			nullParentID = pgtype.UUID{Bytes: pId, Valid: true}

		} else {
			cleanErr := errors.New("parentId query parameter must be a valid UUID")
			app.badRequestResponse(w, r, cleanErr)
			return
		}
	}
	// get the sort parameter
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// the switch
	switch sortOrder {
	case "popular":
		popularComments, dbErr := app.db.ListCommentsByPopular(r.Context(), store.ListCommentsByPopularParams{
			PostID:        postId,
			Limit:         int32(pagination.Limit),
			ParentID:      nullParentID,
			Offset:        pagination.Offset,
			CurrentUserID: currentUserID,
		})
		err = dbErr
		// convert the type
		for _, c := range popularComments {
			comments = append(comments, store.ListCommentsByNewestRow(c))
		}
	case "asc":
		oldestComments, dbErr := app.db.ListCommentsByOldest(r.Context(), store.ListCommentsByOldestParams{
			PostID:        postId,
			Limit:         int32(pagination.Limit),
			ParentID:      nullParentID,
			Cursor:        pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
			CurrentUserID: currentUserID,
		})
		err = dbErr
		// convert the type
		for _, c := range oldestComments {
			comments = append(comments, store.ListCommentsByNewestRow(c))
		}
	case "desc":
		comments, err = app.db.ListCommentsByNewest(r.Context(), store.ListCommentsByNewestParams{
			PostID:   postId,
			Limit:    int32(pagination.Limit),
			ParentID: nullParentID,
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
			CurrentUserID: currentUserID,
		})
	}

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch comments: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// preventing null to the frontend, return empty array instead
	if comments == nil {
		comments = []store.ListCommentsByNewestRow{}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {

	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	updatedComment, err := app.db.UpdateComment(r.Context(), store.UpdateCommentParams{
		ID:        commentId,
		UserID:    userID,
		Content:   req.Content,
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		// comment does not exist or user does not own it
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}

		wrappedErr := fmt.Errorf("failed to update comment in database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"comment": updatedComment})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {

	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	rowsAffected, err := app.db.DeleteComment(r.Context(), store.DeleteCommentParams{
		ID:     commentId,
		UserID: userID,
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to delete comment: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if rowsAffected == 0 {
		app.notFoundResponse(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) voteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	var req VoteCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Vote != 1 && req.Vote != -1) {
		cleanErr := errors.New("Vote must be 1 or -1")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	tx, err := app.pool.Begin(r.Context())
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to start transaction: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}
	// safely abort if no explicit commit
	defer tx.Rollback(r.Context())

	// attach transaction to sqlc queries
	qtx := app.db.WithTx(tx)

	// check user's current vote in the ledger
	currentVote, err := qtx.GetCommentVote(r.Context(), store.GetCommentVoteParams{
		CommentID: commentId,
		UserID:    userID,
	})

	var delta int32 = 0

	// ledger math
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// user never vote on this post
			delta = int32(req.Vote)
			err = qtx.SetCommentVote(r.Context(), store.SetCommentVoteParams{
				CommentID: commentId,
				UserID:    userID,
				Vote:      req.Vote,
			})
		} else {
			wrappedErr := fmt.Errorf("database error: %w", err)
			app.serverErrorResponse(w, r, wrappedErr)
			return
		}
	} else {
		if currentVote == req.Vote {
			// user click the same button, toggling vote off
			delta = -int32(req.Vote)
			err = qtx.RemoveCommentVote(r.Context(), store.RemoveCommentVoteParams{
				CommentID: commentId,
				UserID:    userID,
			})
		} else {
			// user flipped the vote, e.g. upvote to downvote
			delta = int32(req.Vote * 2)
			err = qtx.SetCommentVote(r.Context(), store.SetCommentVoteParams{
				CommentID: commentId,
				UserID:    userID,
				Vote:      req.Vote,
			})
		}
	}

	if err != nil {
		wrappedErr := fmt.Errorf("failed to update vote ledger: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = qtx.UpdateCommentScore(r.Context(), store.UpdateCommentScoreParams{
		ID:    commentId,
		Score: delta,
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to update comment score: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// commit the transaction
	if err = tx.Commit(r.Context()); err != nil {
		wrappedErr := fmt.Errorf("failed to commit transaction: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
