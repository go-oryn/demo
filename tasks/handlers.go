package tasks

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// NewListHandler returns the GET /api/tasks handler.
func NewListHandler(repository *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := repository.List(r.Context())
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusOK, tasks)
	}
}

// NewGetHandler returns the GET /api/tasks/{id} handler.
func NewGetHandler(repository *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := repository.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusOK, task)
	}
}

// createRequest is the body of POST /api/tasks.
type createRequest struct {
	Title string `json:"title"`
}

// NewCreateHandler returns the POST /api/tasks handler.
func NewCreateHandler(repository *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || strings.TrimSpace(request.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": `expected a JSON body like {"title": "..."}`})

			return
		}

		task, err := repository.Create(r.Context(), strings.TrimSpace(request.Title))
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusCreated, task)
	}
}

// NewCompleteHandler returns the POST /api/tasks/{id}/complete handler.
func NewCompleteHandler(repository *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := repository.Complete(r.Context(), r.PathValue("id")); err != nil {
			writeError(w, r, logger, err)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// NewDeleteHandler returns the DELETE /api/tasks/{id} handler.
func NewDeleteHandler(repository *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := repository.Delete(r.Context(), r.PathValue("id")); err != nil {
			writeError(w, r, logger, err)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeError answers 404 for missing tasks, and logs any other error before answering 500.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})

		return
	}

	logger.ErrorContext(r.Context(), "tasks request failed", "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
