// Package middlewares contains middlewares
package middlewares

import (
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// RequestID is a middleware that injects a request ID into the context of each request. A request ID is a string of the form "host.example.com/random-0001", where "random" is a base62 random string that uniquely identifies this go process, and where the last number is an atomically incremented request counter.
func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results of parsing either the True-Client-IP, X-Real-IP or the X-Forwarded-For headers (in that order).
// This middleware should be inserted fairly early in the middleware stack to ensure that subsequent layers (e.g., request loggers) which examine the RemoteAddr will see the intended value.
// You should only use this middleware if you can trust the headers passed to you (in particular, the two headers this middleware uses), for example because you have placed a reverse proxy like HAProxy or nginx in front of chi. If your reverse proxies are configured to pass along arbitrary header values from the client, or if you use this middleware without a reverse proxy, malicious clients will be able to make you very sad (or, depending on how you're using RemoteAddr, vulnerable to an attack of some sort).
func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}

// Logger is a middleware that logs the start and end of each request, along with some useful data about what was requested, what the response status was, and how long it took to return. When standard output is a TTY, Logger will print in color, otherwise it will print in black and white. Logger prints a request ID if one is provided.
//
// Alternatively, look at [https://github.com/goware/httplog](https://github.com/goware/httplog) for a more in-depth http logger with structured logging support.
//
// IMPORTANT NOTE: Logger should go before any other middleware that may change the response, such as middleware.Recoverer. Example:
//
//	r := chi.NewRouter()
//	r.Use(middleware.Logger)        // <--<< Logger should come before Recoverer
//	r.Use(middleware.Recoverer)
//	r.Get("/", handler)
//
// [`middleware.Logger` on pkg.go.dev](https://pkg.go.dev/github.com/go-chi/chi/v5@v5.0.12/middleware#Logger)
func Logger(next http.Handler) http.Handler {
	return middleware.Logger(next)
}

// Recoverer is a middleware that recovers from panics, logs the panic (and a backtrace), and returns a HTTP 500 (Internal Server Error) status if possible. Recoverer prints a request ID if one is provided.
//
// Alternatively, look at [https://github.com/go-chi/httplog](https://github.com/go-chi/httplog) middleware pkgs.
//
// [`middleware.Recoverer` on pkg.go.dev](https://pkg.go.dev/github.com/go-chi/chi/v5@v5.0.12/middleware#Recoverer)
func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

// IsContentJSON is a middleware that checks wether the application content is json
func IsContentJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			log.Error().
				Str("Content-Type", contentType).
				Msg("invalid content type provided")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			sonic.ConfigDefault.NewEncoder(w).Encode(
				map[string]interface{}{
					"message": "only conent type of application/json is allowed",
				})
			return
		}

		next.ServeHTTP(w, r)
	})
}
