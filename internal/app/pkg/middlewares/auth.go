package middlewares

import (
	"context"
	"net/http"
)

// DriverContext contains the driver context
type DriverContext string

// DriverContextKey is the drvier context key that is used to identify the driver in the context
const DriverContextKey DriverContext = "driver_id"

// IsDriver is a middleware that is used to check wether the driver is logged in
func IsDriver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// FIX: Add the logic to extract the drivers JWT token and validate it to check wether the logging in user
		// is infact a driver
		isDriver := true
		if !isDriver {
			http.Error(w, "you are unauthorized to perform this action", http.StatusUnauthorized)
			return
		}

		driverID := 2

		ctx := context.WithValue(r.Context(), DriverContextKey, driverID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
