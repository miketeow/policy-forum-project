package main

import (
	"net/http"
)

func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// return empty array if no query, instead of error
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	results, err := app.db.GlobalSearch(r.Context(), query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
