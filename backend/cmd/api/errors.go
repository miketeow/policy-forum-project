package main

import "net/http"

// --- TIER 1: BASE HELPER ---

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	// wrap the error in enveloper for consistent {"error":"..."} formatting
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env)
	if err != nil {
		// If failed to write the error JSON, fallback to raw HTTP 500
		app.logger.Printf("error writing JSON response for %s %s: %s", r.Method, r.URL.Path, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// --- TIER 2: SEMANTIC WRAPPER ---

// serverErrorResponse(500) (developer's fault)
// log the exact error to the terminal, but hide the detail from the user
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Printf("request error: %s %s: %s", r.Method, r.URL.Path, err.Error())
	message := "the server encounter a problem and count not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// badRequestResponse(400) (user's fault)
// send the specific error detail to the user, but does not clutter the server logs
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// notFoundResponse(404) (strict routing)
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resources could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// invalidAuthenticationTokenResponse(401) (missing or invalid token)
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// invalidCredentialsResponse(401) (use for login failure)
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid email or password"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}
