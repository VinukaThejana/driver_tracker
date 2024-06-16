package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

type (
	// BookingTokenIDCtx contains the type for the booking token ID
	BookingTokenIDCtx string
	// DriverIDCtx contians the type for the driver ID
	DriverIDCtx string
	// BookingIDCtx contains the type for the booking ID type
	BookingIDCtx string
	// PartitionNoCtx contains the type for the partition no type
	PartitionNoCtx string
)

const (
	// BookingTokenID is a key to notate the booking token id
	BookingTokenID BookingTokenIDCtx = "booking_token_id"
	// DriverID is a key to notate the driver id
	DriverID DriverIDCtx = "driver_id"
	// BookingID is a key to notate the booking id
	BookingID BookingIDCtx = "booking_id"
	// PartitionNo is a key to notate the partition number
	PartitionNo PartitionNoCtx = "partition_no"
)

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

		ctx := context.WithValue(r.Context(), DriverID, driverID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsBookingTokenValid is a middleware that is used to check wether the booking token is valid or not
func IsBookingTokenValid(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookingToken := ""
		unauthorizedErr := fmt.Errorf("you are not authorized to perform this operation")

		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) == 2 {
			bookingToken = authorization[1]
		} else {
			bookingTokenC, err := r.Cookie("booking_token")
			if err != nil {
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			bookingToken = bookingTokenC.Value
		}

		if bookingToken == "" {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		bt := tokens.BookingToken{
			C: c,
			E: e,
		}
		isValid, token := bt.Validate(r.Context(), bookingToken)
		if !isValid {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		id, driverID, bookingID, partitionNo, err := bt.Get(token)
		if err != nil {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		ctx = context.WithValue(ctx, BookingTokenID, id)
		ctx = context.WithValue(ctx, DriverID, driverID)
		ctx = context.WithValue(ctx, BookingID, bookingID)
		ctx = context.WithValue(ctx, PartitionNo, partitionNo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
