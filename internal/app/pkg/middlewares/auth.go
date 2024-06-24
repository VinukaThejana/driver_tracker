package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
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
func IsDriver(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driverToken := ""
		unauthorizedErr := errors.ErrUnauthorized

		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) == 2 {
			driverToken = authorization[1]
		} else {
			driverTokenC, err := r.Cookie("EncryptKey")
			if err != nil {
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			driverToken = driverTokenC.Value
		}

		if driverToken == "" {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		dt := tokens.DriverToken{
			E: e,
			C: c,
		}

		isValid, _ := dt.Validate(driverToken)
		if !isValid {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		// FIX: Get the proper driver ID
		ctx := context.WithValue(r.Context(), DriverID, 2)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsBookingTokenValid is a middleware that is used to check wether the booking token is valid or not
func IsBookingTokenValid(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookingToken := ""
		unauthorizedErr := errors.ErrUnauthorized

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

// IsSuperAdmin is a middleware that is used to make sure that the requesting user is the admin
func IsSuperAdmin(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := r.URL.Query().Get("secret")
		if secret != e.AdminSecret {
			http.Error(w, errors.ErrNotAdmin.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
