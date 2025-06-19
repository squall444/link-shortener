package middleware

import "net/http"

type WraperWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WraperWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.StatusCode = statusCode
}
