package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"policy-forum-backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreatePostRequest struct {
	Title    string `json:"title" validate:"required,min=5,max=300"`
	Content  string `json:"content" validate:"required,min=10,max=40000"`
	Category string `json:"category" validate:"omitempty,oneof=INFRASTRUCTURE ECONOMY HEALTHCARE EDUCATION ENVIRONMENT SAFETY OTHER"`
}

type UpdatePostRequest struct {
	Title    string `json:"title" validate:"required,min=5,max=300"`
	Content  string `json:"content" validate:"required,min=10,max=40000"`
	Category string `json:"category" validate:"omitempty,oneof=INFRASTRUCTURE ECONOMY HEALTHCARE EDUCATION ENVIRONMENT SAFETY OTHER"`
}

type VotePostRequest struct {
	Vote int16 `json:"vote" validate:"required,oneof=1 -1"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
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
	// execute the insert
	post, err := app.db.CreatePost(r.Context(), args)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to create post in database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	bgCtx := context.WithoutCancel(r.Context())

	// new go routine in background, to create correct category for the post
	go func(bgCtx context.Context, postID uuid.UUID, postTitle, postContent string) {
		// call gemini
		aiCategory := app.categorizeWithAI(bgCtx, postTitle, postContent)
		app.LogInfo(bgCtx, "ai categorized post", slog.String("post_id", postID.String()), slog.String("category", aiCategory))

		// update the database
		// use context.Background() because the original HTTP request context
		// will be cancelled the moment the user gets their HTTP response
		err := app.db.UpdatePostCategory(bgCtx, store.UpdatePostCategoryParams{
			ID:       postID,
			Category: store.PostCategory(aiCategory),
		})
		if err != nil {
			app.LogError(bgCtx, "failed to update post with category",
				slog.String("post_id", postID.String()),
				slog.String("error", err.Error()))

		}
	}(bgCtx, post.ID, post.Title, post.Content) // pass variables here

	err = app.writeJSON(w, http.StatusCreated, envelope{"post": post})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listPostHandler(w http.ResponseWriter, r *http.Request) {
	var posts []store.ListPostsByNewestRow
	var err error

	pagination, err := app.parsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	hasCursor := !pagination.Cursor.IsZero()

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	var currentUserID pgtype.UUID

	if userID, ok := r.Context().Value(userIDKey).(uuid.UUID); ok {
		currentUserID = pgtype.UUID{Bytes: userID, Valid: true}
	}

	switch sortOrder {
	case "popular":
		// define the cache key
		cacheKey := fmt.Sprintf("posts:popular:anon:limit:%d:offset:%d", pagination.Limit, pagination.Offset)

		// for unauthenticated user
		if !currentUserID.Valid {
			// get posts from redis
			cachedPosts, err := app.rdb.Get(r.Context(), cacheKey).Result()
			if err == nil {
				// hit the cache, deserialise directly into the struct slice
				app.LogInfo(r.Context(), "redis cache hit", slog.String("key", cacheKey))
				if unmarshalErr := json.Unmarshal([]byte(cachedPosts), &posts); unmarshalErr == nil {
					break
				}
			}
		}

		// cache miss or authenticated user, get posts from database
		app.LogInfo(r.Context(), "database hit", slog.String("key", cacheKey))

		popularPosts, dbErr := app.db.ListPostsByPopular(r.Context(), store.ListPostsByPopularParams{
			Limit:         int32(pagination.Limit),
			Offset:        pagination.Offset,
			CurrentUserID: currentUserID,
		})
		err = dbErr

		for _, p := range popularPosts {
			posts = append(posts, store.ListPostsByNewestRow(p))
		}

		// save to redis if it is an unauthenticated query
		if !currentUserID.Valid && err == nil {
			if jsonBytes, marshalErr := json.Marshal(posts); marshalErr == nil {
				app.rdb.Set(r.Context(), cacheKey, jsonBytes, time.Hour)
			}
		}
	case "asc":
		oldestPost, dbErr := app.db.ListPostsByOldest(r.Context(), store.ListPostsByOldestParams{
			Limit: int32(pagination.Limit),
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
			CurrentUserID: currentUserID,
		})
		err = dbErr

		for _, p := range oldestPost {
			posts = append(posts, store.ListPostsByNewestRow(p))
		}
	case "desc":
		posts, err = app.db.ListPostsByNewest(r.Context(), store.ListPostsByNewestParams{
			Limit: int32(pagination.Limit),
			Cursor: pgtype.Timestamp{
				Time:  pagination.Cursor,
				Valid: hasCursor,
			},
			CurrentUserID: currentUserID,
		})
	}

	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch posts from database: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}
	// if no posts exist, make sure to return empty array instead of null
	if posts == nil {
		posts = []store.ListPostsByNewestRow{}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"posts": posts})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	// get user id, it can be empty if not logged in
	var currentUserID pgtype.UUID

	if userID, ok := r.Context().Value(userIDKey).(uuid.UUID); ok {
		currentUserID = pgtype.UUID{Bytes: userID, Valid: true}
	}

	// extract id from params
	idParam := r.PathValue("postId")

	postId, err := uuid.Parse(idParam)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// fetch from database
	post, err := app.db.GetPostByID(r.Context(), store.GetPostByIDParams{
		ID:            postId,
		CurrentUserID: currentUserID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}
		wrappedErr := fmt.Errorf("failed to fetch post by id: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		wrappedErr := fmt.Errorf("critical error: user id missing from context")
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	postIDParam := r.PathValue("postId")
	postId, err := uuid.Parse(postIDParam)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cleanErr := errors.New("the provided JSON payload is malformed or invalid")
		app.badRequestResponse(w, r, cleanErr)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	category := req.Category

	if category == "" {
		category = "PENDING"
	}

	updatedPost, err := app.db.UpdatePost(r.Context(), store.UpdatePostParams{
		ID:        postId,
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Category:  store.PostCategory(category),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		// if wrong id or wrong users
		if errors.Is(err, pgx.ErrNoRows) {
			app.notFoundResponse(w, r)
			return
		}

		// if database crash
		wrappedErr := fmt.Errorf("failed to update post: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// detach context
	bgCtx := context.WithoutCancel(r.Context())

	// background task: re-categorized modified content
	go func(bgCtx context.Context, postID uuid.UUID, postTitle, postContent string) {
		aiCategory := app.categorizeWithAI(bgCtx, postTitle, postContent)
		app.LogInfo(bgCtx, "ai re-categorized post",
			slog.String("post_id", postID.String()),
			slog.String("category", aiCategory))

		err := app.db.UpdatePostCategory(bgCtx, store.UpdatePostCategoryParams{
			ID:       postID,
			Category: store.PostCategory(aiCategory),
		})
		if err != nil {
			app.LogError(bgCtx, "failed to update post with category",
				slog.String("post_id", postID.String()),
				slog.String("error", err.Error()))
		}
	}(bgCtx, updatedPost.ID, updatedPost.Title, updatedPost.Content)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": updatedPost})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
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

	rowsAffected, err := app.db.DeletePost(r.Context(), store.DeletePostParams{
		ID:     postId,
		UserID: userID,
	})

	if err != nil {
		// if database crash
		wrappedErr := fmt.Errorf("failed to delete post: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if rowsAffected == 0 {
		app.notFoundResponse(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) votePostHandler(w http.ResponseWriter, r *http.Request) {
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

	var req VotePostRequest
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
	defer tx.Rollback(context.Background())

	// attach transaction to sqlc queries
	qtx := app.db.WithTx(tx)

	// check user's current vote in the ledger
	currentVote, err := qtx.GetPostVoteForUpdate(r.Context(), store.GetPostVoteForUpdateParams{
		PostID: postId,
		UserID: userID,
	})

	var delta int32 = 0

	// ledger math
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// user never vote on this post
			delta = int32(req.Vote)
			err = qtx.InsertPostVote(r.Context(), store.InsertPostVoteParams{
				PostID: postId,
				UserID: userID,
				Vote:   req.Vote,
			})
			// handle concurrency
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					app.errorResponse(w, r, http.StatusConflict, "Vote already processing")
					return
				}
				app.serverErrorResponse(w, r, fmt.Errorf("failed to insert vote: %w", err))
				return
			}
		} else {
			wrappedErr := fmt.Errorf("database error: %w", err)
			app.serverErrorResponse(w, r, wrappedErr)
			return
		}
	} else {
		// user already vote
		if currentVote == req.Vote {
			// user click the same button, toggling vote off
			delta = -int32(req.Vote)
			err = qtx.RemovePostVote(r.Context(), store.RemovePostVoteParams{
				PostID: postId,
				UserID: userID,
			})
		} else {
			// user flipped the vote, e.g. upvote to downvote
			delta = int32(req.Vote * 2)
			err = qtx.UpdatePostVote(r.Context(), store.UpdatePostVoteParams{
				PostID: postId,
				UserID: userID,
				Vote:   req.Vote,
			})
		}
	}

	if err != nil {
		wrappedErr := fmt.Errorf("failed to update vote ledger: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	err = qtx.AtomicUpdatePostScore(r.Context(), store.AtomicUpdatePostScoreParams{
		ID:    postId,
		Score: delta,
	})

	if err != nil {
		wrappedErr := fmt.Errorf("failed to update post score: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	// commit the transaction
	if err = tx.Commit(r.Context()); err != nil {
		wrappedErr := fmt.Errorf("failed to commit transaction: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	go func() {
		ctx := context.Background()
		app.rdb.Del(ctx, "posts:popular:anon:limit:10:offset:0")
		app.LogInfo(ctx, "redis cache invalidated", slog.String("event", "vote_recorded"))
	}()

	w.WriteHeader(http.StatusOK)
}
