// Package middlewares contains middlewares
package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// CheckContentIsJSON is a middleware that checks wether the application content is json
func CheckContentIsJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			log.Error().Str("Content-Type", contentType).Msg("invalid content type provided")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "only conent type of application/json is allowed",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
