package main

import (
	"fmt"
	"net/http"
	"policy-forum-backend/internal/store"
)

func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// return empty array if no query, instead of error
		err := app.writeJSON(w, http.StatusOK, envelope{"result": []any{}})
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	results, err := app.db.GlobalSearch(r.Context(), query)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch any search result: %w", err)
		app.serverErrorResponse(w, r, wrappedErr)
		return
	}

	if results == nil {
		results = []store.GlobalSearchRow{}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"result": results})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
