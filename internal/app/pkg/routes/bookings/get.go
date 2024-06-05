package bookings

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

// getBookingByID is a route that is used to get the booking information by the booking id
func getBookingByID(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
}
