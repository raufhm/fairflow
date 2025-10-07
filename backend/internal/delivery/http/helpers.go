package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// getIDFromPath extracts an ID from the URL path
func getIDFromPath(r *http.Request, prefix string, suffixes ...string) int64 {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	for _, suffix := range suffixes {
		path = strings.TrimSuffix(path, suffix)
	}
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parseID(parts[0])
	}
	return 0
}

// parseID converts a string to int64
func parseID(s string) int64 {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// getCurrentTimestamp returns current timestamp in ISO format
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}