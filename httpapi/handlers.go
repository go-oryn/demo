package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-oryn/demo/tasks"
)

// NewListHandler returns the GET /api/tasks handler.
func NewListHandler(service *tasks.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := service.List(r.Context())
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusOK, list)
	}
}

// NewGetHandler returns the GET /api/tasks/{id} handler.
func NewGetHandler(service *tasks.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := service.Get(r.Context(), r.PathValue("id"))
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
func NewCreateHandler(service *tasks.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody(`expected a JSON body like {"title": "..."}`))

			return
		}

		task, err := service.Create(r.Context(), request.Title)
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusCreated, task)
	}
}

// NewCompleteHandler returns the POST /api/tasks/{id}/complete handler.
func NewCompleteHandler(service *tasks.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := service.Complete(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, r, logger, err)

			return
		}

		writeJSON(w, http.StatusOK, task)
	}
}

// NewDeleteHandler returns the DELETE /api/tasks/{id} handler.
func NewDeleteHandler(service *tasks.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := service.Delete(r.Context(), r.PathValue("id")); err != nil {
			writeError(w, r, logger, err)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func errorBody(message string) map[string]string {
	return map[string]string{"error": message}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeError maps the tasks errors to HTTP statuses, and logs the unexpected ones.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody(err.Error()))
	case errors.Is(err, tasks.ErrInvalidTitle):
		writeJSON(w, http.StatusBadRequest, errorBody(err.Error()))
	default:
		logger.ErrorContext(r.Context(), "tasks request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal error"))
	}
}
