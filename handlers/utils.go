package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	contentTypeHTML = "text/html"
)

func sendError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

func setHTMLHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentTypeHTML)
}
