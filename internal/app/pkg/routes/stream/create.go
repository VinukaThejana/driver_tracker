package stream

import (
	"context"
	"database/sql"
	ers "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// Create is a route that is used to create a new stream for the given booking id
func create(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const (
		// Below variables are used to lock the partition manager until a proper partition is chosen
		// for the job creation
		//
		// The time for which the partition manager will be locked
		lockDuration = 2 * time.Second
		// The maximum time that another request will be wating till the partition manager is freed
		// before timeout
		timeout = 3 * time.Second

		maxRequestBodySize = 1 << 6
	)

	type body struct {
		BookingID string `json:"booking_id" validate:"required,min=1"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).
				Msg("failed to read the request body")
			lib.JSONResponse(w, http.StatusRequestEntityTooLarge, errors.ErrBookingIDNotValid.Error())
			return
		}

		log.Error().Err(err).
			Msgf(
				"raw_body : %s\tfailed to read the request body",
				string(body),
			)
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	var driverID *int
	var bookPickupAddr *string

	query := `
SELECT
	DriverPk,
  BookPickUpAddr
FROM
	Tbl_BookingDetails
WHERE
	BookRefNo = @BookRefNo;
`

	err = c.DB.QueryRow(query, sql.Named("BookRefNo", reqBody.BookingID)).Scan(&driverID, &bookPickupAddr)
	if err != nil || driverID == nil {
		log.Error().Err(err).
			Msgf(
				"booking_id : %s\tfailed to get the driver ID from the booking ID",
				reqBody.BookingID,
			)
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	pickups := []services.Geo{}
	if bookPickupAddr != nil {
		pickups = append(pickups, services.Geocode(r.Context(), e, c, true, lib.Seperator(*bookPickupAddr, "|"))...)
	}
	if len(pickups) == 0 {
		log.Error().Msg("cannot find the pickup address for the given location")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	client := c.R.DB

	if val := client.Get(r.Context(), fmt.Sprint(*driverID)).Val(); val != "" {
		DriverID := _lib.NewDriverID()
		err = sonic.UnmarshalString(val, &DriverID)
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"driver_id : %d\tredis_value : %s\tbooking_id : %s\tfailed to unmarshal",
					*driverID,
					val,
					reqBody.BookingID,
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}

		bookingID := DriverID[_lib.DriverIDBookingID]
		partition, err := strconv.Atoi(DriverID[_lib.DriverIDPartitionNo])
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"driver_id : %d\tredis_value : %s\tbooking_id : %s\tpartition : %s\tfailed to unmarshal",
					*driverID,
					val,
					reqBody.BookingID,
					DriverID[_lib.DriverIDPartitionNo],
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}

		if bookingID == reqBody.BookingID {
			token, ttl, err := renewBookingToken(
				r.Context(),
				e,
				c,
				*driverID,
				partition,
				reqBody.BookingID,
			)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"booking_id : %s\tdriver_id : %d\tfailed to renew the booking token",
						bookingID,
						*driverID,
					)
				lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "booking_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   e.BookingTokenExpires,
				Expires:  time.Now().UTC().Add(ttl).UTC(),
			})

			lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
				"booking_token": token,
			})
			return
		}

		val = client.Get(r.Context(), _lib.N(partition)).Val()
		if val == "" {
			log.Error().Err(err).
				Msgf(
					"booking_id : %s\tdriver_id : %d\tfailed to get the n- partition from redis",
					bookingID,
					*driverID,
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}
		N := _lib.NewN()
		err = sonic.UnmarshalString(val, &N)
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"booking_id : %s\tdriver_id : %d\tfailed to parse n- partition",
					bookingID,
					*driverID,
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}
		offset, err := strconv.Atoi(N[_lib.NLastOffset])
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"booking_id : %s\tdriver_id : %d\tredis_value : %s\tfailed to conver the offset to int",
					bookingID,
					*driverID,
					val,
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}

		err = _lib.DelBooking(r.Context(), client, *driverID, bookingID, partition)
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"booking_id : %s\tdriver_id : %d\tfailed to delete the keys in redis",
					bookingID,
					*driverID,
				)
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}

		go services.GenerateLog(
			e,
			c,
			bookingID,
			partition,
			int64(offset),
		)
	}

	var available []int

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	serverBusyErr := fmt.Errorf("server is busy right now, no partitions are currently available")

	err = func() error {
		acquired, err := c.R.AcquireLock(ctx, client, e.PartitionManagerKey, lockDuration, 100*time.Millisecond)
		if err != nil {
			return err
		}
		if !acquired {
			return serverBusyErr
		}
		defer c.R.ReleaseLock(client, e.PartitionManagerKey)

		payload := client.SMembers(r.Context(), e.PartitionManagerKey).Val()
		jobs := make(map[int]struct{})
		for _, data := range payload {
			job, err := strconv.Atoi(data)
			if err != nil {
				client.SRem(r.Context(), e.PartitionManagerKey, data)
				return err
			}
			jobs[job] = struct{}{}
		}
		partitions := make([]int, e.TotalPartitions)
		for i := range partitions {
			partitions[i] = i
		}
		for _, partition := range partitions {
			_, found := jobs[partition]
			if !found {
				available = append(available, partition)
			}
		}

		if len(available) == 0 {
			return serverBusyErr
		}

		err = client.SAdd(r.Context(), e.PartitionManagerKey, available[0]).Err()
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		if ers.Is(err, serverBusyErr) {
			lib.JSONResponse(w, http.StatusConflict, errors.ErrBusy.Error())
			return
		}

		log.Error().Err(err)
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	partition := available[0]

	lastOffset, err := c.GetLastOffset(r.Context(), e, e.Topic, partition)
	if err != nil {
		log.Error().Err(err).Msg("failed to get the lastoffset")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	newOffset := int(lastOffset) + 1

	bt := tokens.NewBookingToken(e, c)

	token, err := bt.Create(r.Context(), tokens.BookingTokenOpts{
		DriverID:         *driverID,
		BookingID:        reqBody.BookingID,
		Partition:        partition,
		NewOffset:        newOffset,
		PickupCordinates: pickups[0],
	})
	if err != nil {
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", e.BookingTokenExpires))
	if err != nil {
		log.Error().Err(err).Msg("failed to convert the booking token expires to seconds")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "booking_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   e.BookingTokenExpires,
		Expires:  time.Now().UTC().Add(duration).UTC(),
	})

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"booking_token": token,
	})
}

func renewBookingToken(
	ctx context.Context,
	e *env.Env,
	c *connections.C,
	driverID,
	partition int,
	bookingID string,
) (token string, ttl time.Duration, err error) {
	client := c.R.DB

	ttl = client.TTL(ctx, fmt.Sprint(driverID)).Val()
	if ttl <= 0 {
		return "", 0, fmt.Errorf("the ttl the previous booking token is smaller than 0")
	}

	bt := tokens.NewBookingToken(e, c)

	id, token, err := bt.Createtoken(
		driverID,
		partition,
		bookingID,
		ttl,
	)
	if err != nil {
		return "", 0, err
	}

	driverDetails, err := sonic.MarshalString(_lib.SetDriverID(id.String(), bookingID, partition))
	if err != nil {
		return "", 0, err
	}

	client.Set(ctx, fmt.Sprint(driverID), driverDetails, ttl)

	return token, 0, nil
}
