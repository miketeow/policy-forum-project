package main

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// --- SYSTEM & GLOBAL ROUTES ---
	mux.HandleFunc("GET /api/search", app.searchHandler)

	// --- AUTHENTICATION ---
	mux.HandleFunc("POST /api/auth/register", app.registerHandler)
	mux.HandleFunc("POST /api/auth/login", app.loginHandler)
	mux.HandleFunc("POST /api/auth/logout", app.requireAuth(app.logoutHandler))

	// --- USER PROFILE & DASHBOARD
	mux.HandleFunc("GET /api/users/me", app.requireAuth(app.getUserProfileHandler))
	mux.HandleFunc("GET /api/users/me/posts", app.requireAuth(app.getUserPostsHandler))
	mux.HandleFunc("GET /api/users/me/comments", app.requireAuth(app.getUserCommentsHandler))
	mux.HandleFunc("GET /api/users/me/upvoted/posts", app.requireAuth(app.getUserUpvotedPostsHandler))
	mux.HandleFunc("GET /api/users/me/upvoted/comments", app.requireAuth(app.getUserUpvotedCommentsHandler))

	// --- POSTS (Core Resource)
	// Public / Optional Auth (Read operation)
	mux.HandleFunc("GET /api/posts", app.optionalAuth(app.listPostHandler))
	mux.HandleFunc("GET /api/posts/{postId}", app.optionalAuth(app.getPostHandler))
	mux.HandleFunc("GET /api/posts/{postId}/comments", app.optionalAuth(app.getCommentsHandler))

	// Protected (Write operation)
	mux.HandleFunc("POST /api/posts", app.requireAuth(app.createPostHandler))
	mux.HandleFunc("PUT /api/posts/{postId}", app.requireAuth(app.updatePostHandler))
	mux.HandleFunc("DELETE /api/posts/{postId}", app.requireAuth(app.deletePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/vote", app.requireAuth(app.votePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/comments", app.requireAuth(app.createCommentHandler))

	// -- COMMENTS (Independent Actions)
	// Protected (Write operation)
	mux.HandleFunc("PUT /api/comments/{commentId}", app.requireAuth(app.updateCommentHandler))
	mux.HandleFunc("DELETE /api/comments/{commentId}", app.requireAuth(app.deleteCommentHandler))
	mux.HandleFunc("POST /api/comments/{commentId}/vote", app.requireAuth(app.voteCommentHandler))

	// Background AI Endpoint
	mux.HandleFunc("GET /api/posts/{postID}/stream", app.streamPostStatusHandler)
	mux.HandleFunc("POST /api/posts/{postID}/summary", app.requireAuth(app.triggerSummaryHandler))

	// Category Report Endpoint
	mux.HandleFunc("GET /api/reports/{category}", app.getCategoryReportHandler)
	mux.HandleFunc("POST /api/reports/{category}/generate", app.requireAuth(app.triggerCategoryReportHandler))
	mux.HandleFunc("GET /api/reports/{category}/status", app.getReportStatusHandler)

	handler := corsMiddleware(mux)

	return otelhttp.NewHandler(handler, "policy-forum-api")
}
