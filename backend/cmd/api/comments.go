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
	if sortOrder == "asc" {
		oldestComments, dbErr := app.db.ListCommentsByOldest(r.Context(), store.ListCommentsByOldestParams{
			PostID:   postId,
			Limit:    int32(pagination.Limit),
			ParentID: nullParentID,
			Cursor:   pgtype.Timestamp{Time: pagination.Cursor, Valid: hasCursor},
		})
		err = dbErr
		// convert the type
		for _, c := range oldestComments {
			comments = append(comments, store.ListCommentsByNewestRow(c))
		}
	} else {

		// call db method
		comments, err = app.db.ListCommentsByNewest(r.Context(), store.ListCommentsByNewestParams{
			PostID:   postId,
			Limit:    int32(pagination.Limit),
			ParentID: nullParentID,
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
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
