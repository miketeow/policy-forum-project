package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCommentRequest struct {
	Content  string     `json:"content"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type VoteCommentRequest struct {
	Vote int16 `json:"vote"`
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

	// if req.ParentID != nil {
	// 	log.Printf("[CREATE TRACER] Saving Reply! ParentID: %s | Content: %s", req.ParentID.String(), req.Content)
	// } else {
	// 	log.Printf("[CREATE TRACER] Saving Root Comment! ParentID is NIL | Content: %s", req.Content)
	// }

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
	// declare master slice
	var comments []store.ListCommentsByNewestRow
	var err error

	var currentUserID pgtype.UUID

	if userID, ok := r.Context().Value(userIDKey).(uuid.UUID); ok {
		currentUserID = pgtype.UUID{Bytes: userID, Valid: true}
	}

	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid post ID format")
		return
	}

	// pagination
	pagination := parsePagination(r)
	hasCursor := !pagination.Cursor.IsZero()

	// look for optional parent ID
	var nullParentID pgtype.UUID
	parentIDStr := r.URL.Query().Get("parentId")
	if parentIDStr != "" {
		pId, err := uuid.Parse(parentIDStr)
		if err == nil {
			nullParentID = pgtype.UUID{Bytes: pId, Valid: true}
			// log.Printf("[FETCH TRACER] Fetching Replies for Parent: %s", parentIDStr)
		} else {
			// log.Printf("[FETCH TRACER ERROR] Invalid UUID from frontend: '%s'", parentIDStr)
			writeJSONError(w, http.StatusBadRequest, "Invalid parent ID format")
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
		log.Printf("Failed to fetch comments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// preventing null to the frontend, return empty array instead
	if comments == nil {
		comments = []store.ListCommentsByNewestRow{}
	}

	writeJSON(w, http.StatusOK, comments)
}
func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {

	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid comment ID format")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	updatedComment, err := app.db.UpdateComment(r.Context(), store.UpdateCommentParams{
		ID:        commentId,
		UserID:    userID,
		Content:   req.Content,
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		log.Printf("Failed to update comment: %v", err)
		writeJSONError(w, http.StatusForbidden, "Not authorized to edit this comment, or comment does not exist")
		return
	}

	writeJSON(w, http.StatusOK, updatedComment)
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {

	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid comment ID format")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = app.db.DeleteComment(r.Context(), store.DeleteCommentParams{
		ID:     commentId,
		UserID: userID,
	})

	if err != nil {
		log.Printf("Failed to delete comment: %v", err)
		writeJSONError(w, http.StatusForbidden, "Not authorized to delete this comment, or comment does not exist")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) voteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDParam := r.PathValue("commentId")
	commentId, err := uuid.Parse(commentIDParam)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid comment ID format")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req VoteCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Vote != 1 && req.Vote != -1) {
		writeJSONError(w, http.StatusBadRequest, "Vote must be 1 or -1")
		return
	}

	tx, err := app.pool.Begin(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to start transaction")
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
			writeJSONError(w, http.StatusInternalServerError, "Database error")
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
		writeJSONError(w, http.StatusInternalServerError, "Failed to update ledger")
		return
	}

	err = qtx.UpdateCommentScore(r.Context(), store.UpdateCommentScoreParams{
		ID:    commentId,
		Score: delta,
	})

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to update comment score")
		return
	}

	// commit the transaction
	if err = tx.Commit(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusOK)
}
