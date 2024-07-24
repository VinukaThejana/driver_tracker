package middlewares

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	_errors "github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/rs/zerolog/log"
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
	// AdminIDCtx contains the type for the Admin ID
	AdminIDCtx string
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
	// AdminID is a key to indicate the admin id
	AdminID AdminIDCtx = "admin_id"
)

// IsDriver is a middleware that is used to check wether the driver is logged in
func IsDriver(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		driverToken := ""
		unauthorizedErr := _errors.ErrUnauthorized

		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) == 2 {
			driverToken = authorization[1]
		} else {
			driverTokenC, err := r.Cookie(e.DriverCookieName)
			if err != nil {
				log.Error().Err(err).Msg("failed to read the cookie")
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			driverToken = driverTokenC.Value
		}

		if driverToken == "" {
			log.Error().
				Msgf(
					"header : %v\tauthorization : %s\tfailed to authenticate the user",
					r.Header.Clone(),
					r.Header.Get("Authorization"),
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		dt := tokens.NewDriverToken(e, c)

		isValid, token := dt.Validate(driverToken)
		if !isValid {
			log.Error().
				Msgf(
					"driver_token : %s\tfailed to authenticate the driver",
					driverToken,
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}
		driverID, _, err := dt.Get(token)
		if err != nil {
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), DriverID, driverID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsBookingTokenValid is a middleware that is used to check wether the booking token is valid or not
func IsBookingTokenValid(
	next http.Handler,
	e *env.Env,
	c *connections.C,
	validateRemote bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookingToken := ""
		unauthorizedErr := _errors.ErrUnauthorized

		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) == 2 {
			bookingToken = authorization[1]
		} else {
			bookingTokenC, err := r.Cookie(e.BookingCookieName)
			if err != nil {
				log.Error().Err(err).Msg("failed to read the cookie")
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			bookingToken = bookingTokenC.Value
		}

		if bookingToken == "" {
			log.Error().
				Msgf(
					"header : %v\tauthorization : %s\tfailed to authenticate the booking token",
					r.Header.Clone(),
					r.Header.Get("Authorization"),
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		bt := tokens.NewBookingToken(e, c)

		isValid, token := bt.Validate(r.Context(), bookingToken, validateRemote)
		if !isValid {
			log.Error().
				Msgf(
					"booking_token : %s\tfailed to validate the booking token",
					bookingToken,
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		_, driverID, bookingID, partitionNo, err := bt.Get(token)
		if err != nil {
			log.Error().
				Msgf(
					"booking_token : %s\tfailed to validate the booking token",
					bookingToken,
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		ctx = context.WithValue(ctx, DriverID, driverID)
		ctx = context.WithValue(ctx, BookingID, bookingID)
		ctx = context.WithValue(ctx, PartitionNo, partitionNo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateDriverOrBookingToken used to validate the driver or booking token
func ValidateDriverOrBookingToken(
	next http.Handler,
	e *env.Env,
	c *connections.C,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		unauthorizedErr := _errors.ErrUnauthorized

		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) == 2 {
			token = authorization[1]
		}

		if token == "" {
			log.Error().
				Msgf(
					"header: %v\tauthorization : %s\tfailed to get the driver token or the booking token",
					r.Header.Clone(),
					r.Header.Get("Authorization"),
				)
			http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
			return
		}

		bt := tokens.NewBookingToken(e, c)
		dt := tokens.NewDriverToken(e, c)
		isBookingToken := true

		isValid, tk := bt.Validate(r.Context(), token, true)
		if !isValid {
			isValid, tk = dt.Validate(token)
			if !isValid {
				log.Error().
					Msgf("token : %s\tfailed to validate the token", token)
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			isBookingToken = false
		}

		var (
			bookingID   string
			driverID    int
			partitionNo int
			err         error
		)

		if isBookingToken {
			_, driverID, bookingID, partitionNo, err = bt.Get(tk)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"booking_token : %s\tfailed to get the booking token details",
						token,
					)
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}
		} else {
			driverID, _, err = dt.Get(tk)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"driver_token : %s\tfailed to validate the booking token",
						token,
					)
				http.Error(w, unauthorizedErr.Error(), http.StatusUnauthorized)
				return
			}

			client := c.R.DB

			val := client.Get(r.Context(), fmt.Sprint(driverID)).Val()
			if val == "" {
				log.Error().
					Msgf(
						"driver_token : %s\tfailed to get value from Redis",
						token,
					)
				http.Error(w, _errors.ErrServer.Error(), http.StatusInternalServerError)
				return
			}
			payload := make([]string, 3)
			err = sonic.UnmarshalString(val, &payload)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"driver_token : %s\tredis-value : %s\tfailed to parse the redis value",
						token,
						val,
					)
				http.Error(w, _errors.ErrServer.Error(), http.StatusInternalServerError)
				return
			}

			bookingID = payload[1]
			partitionNo, err = strconv.Atoi(payload[2])
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"driver_token : %s\tbooking_id : %s\tpartition : %s\tfailed to convert partition number to int",
						token,
						bookingID,
						payload[2],
					)
				http.Error(w, _errors.ErrServer.Error(), http.StatusInternalServerError)
				return
			}
		}

		ctx := r.Context()

		ctx = context.WithValue(ctx, DriverID, driverID)
		ctx = context.WithValue(ctx, BookingID, bookingID)
		ctx = context.WithValue(ctx, PartitionNo, partitionNo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isSuperAdmin(r *http.Request, e *env.Env, _ *connections.C) (isAdmin bool, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://edge-config.vercel.com/%s/item/secret", e.EdgeConfig), nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+e.EdgeConfigReadToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, _errors.ErrServer
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, _errors.ErrServer
	}

	var payload any
	err = sonic.Unmarshal(body, &payload)
	if err != nil {
		return false, err
	}
	secret, ok := payload.(string)
	if !ok {
		return false, _errors.ErrServer
	}
	key := r.URL.Query().Get("secret")
	if key != secret {
		return false, _errors.ErrUnauthorized
	}

	return true, nil
}

// IsSuperAdmin is a middleware that is used to make sure that the requesting user is the super admin (usually the developer)
func IsSuperAdmin(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isSuperAdmin, err := isSuperAdmin(r, e, c)
		if err != nil {
			if errors.Is(err, _errors.ErrUnauthorized) {
				http.Error(w, _errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			log.Error().Err(err).Msg("failed to validate the superadmin")
			http.Error(w, _errors.ErrServer.Error(), http.StatusInternalServerError)
			return
		}
		if !isSuperAdmin {
			http.Error(w, _errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getAdmin(r *http.Request, e *env.Env, c *connections.C) (adminID int, err error) {
	adminToken := ""
	unauthorizedErr := _errors.ErrUnauthorized

	authorization := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authorization) == 2 {
		adminToken = authorization[1]
	} else {
		adminTokenC, err := r.Cookie(e.AdminCookieName)
		if err != nil {
			log.Error().Err(err).Msg("failed to get the admin cookie")
			return -1, unauthorizedErr
		}

		adminToken = adminTokenC.Value
	}

	at := tokens.NewAdminToken(e, c)

	isValid, token := at.Validate(adminToken)
	if !isValid {
		log.Error().
			Msgf(
				"admin_token : %s\tfailed to validate the admin",
				adminToken,
			)
		return -1, unauthorizedErr
	}

	adminID, err = at.Get(token)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate the admin token")
		return -1, unauthorizedErr
	}

	return adminID, nil
}

// IsAdminOrIsSuperAdmin is a middleware that is used to check wether the requesting user is the super admin or the admin
func IsAdminOrIsSuperAdmin(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isSuperAdmin, _ := isSuperAdmin(r, e, c)
		if isSuperAdmin {
			next.ServeHTTP(w, r)
			return
		}

		_, err := getAdmin(r, e, c)
		if err != nil {
			http.Error(w, _errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IsAdmin is a middleware that is used to make sure that the requesting user is a spoton admin
func IsAdmin(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminID, err := getAdmin(r, e, c)
		if err != nil {
			http.Error(w, _errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, AdminID, fmt.Sprint(adminID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsCron is a middleware that is used to make sure that the requesting entity is a cronjob
func IsCron(next http.Handler, e *env.Env, c *connections.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := r.URL.Query().Get("secret")
		if secret != e.AdminSecret {
			http.Error(w, _errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
