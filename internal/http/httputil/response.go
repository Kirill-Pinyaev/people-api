package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func Error(w http.ResponseWriter, code int, format string, args ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	msg := fmt.Sprintf(format, args...)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   strings.ToLower(http.StatusText(code)),
		"message": msg,
	})
}

func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
