package stream

import (
	"context"
	"database/sql"
	ers "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	lockDuration = 2 * time.Second
	timeout      = 5 * time.Second
)

type body struct {
	BookingID string `json:"booking_id" validate:"required,min=1"`
}

// Create is a route that is used to create a new stream for the given booking id
func create(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	var driverID *int

	query := `
SELECT
	DriverPk
FROM
	Tbl_BookingDetails
WHERE
	BookRefNo = @BookRefNo;
`

	err := c.DB.QueryRow(query, sql.Named("BookRefNo", reqBody.BookingID)).Scan(&driverID)
	if err != nil || driverID == nil {
		log.Error().Err(err).Str("booking_id", reqBody.BookingID).Msg("failed to get the driver ID from the booking ID")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	client := c.R.DB
	ErrBookingAlreadyProcessing := fmt.Errorf("booking id that you provided is already processing, please use another booking id")

	if val := client.Get(r.Context(), reqBody.BookingID).Val(); val != "" {
		isDriver, err := isDriver(c, *driverID, reqBody.BookingID)
		if err != nil || !isDriver {
			log.Error().Err(err).Msg("failed to validate wether the driver owns the booking id")
			lib.JSONResponse(w, http.StatusConflict, ErrBookingAlreadyProcessing.Error())
			return
		}

		token, ttl, err := generate(r.Context(), e, c, *driverID, reqBody.BookingID)
		if err != nil {
			log.Error().Err(err).Msg("failed to renew the booking token")
			lib.JSONResponse(w, http.StatusConflict, ErrBookingAlreadyProcessing.Error())
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

	payload, err := sonic.MarshalString([]int{partition, newOffset, *driverID})
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the interface")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	bt := tokens.NewBookingToken(e, c)

	token, err := bt.Create(r.Context(), *driverID, reqBody.BookingID, partition, newOffset, payload)
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

func isDriver(
	c *connections.C,
	driverID int,
	bookingID string,
) (isDriverOwned bool, err error) {
	var pk *int
	query := "SELECT DriverPk FROM Tbl_BookingDetails WHERE BookRefNo = @BookRefNo"

	err = c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(&pk)
	if err != nil {
		return false, err
	}
	if pk == nil {
		return false, fmt.Errorf("failed to get the primary key of the driver")
	}

	if *pk != driverID {
		return false, fmt.Errorf("the booking id does not belong to the driver")
	}

	return true, nil
}

func generate(
	ctx context.Context,
	e *env.Env,
	c *connections.C,
	driverID int,
	bookingID string,
) (token string, ttl time.Duration, err error) {
	client := c.R.DB

	val := client.Get(ctx, fmt.Sprint(driverID)).Val()
	if val == "" {
		return "", 0, fmt.Errorf("the driver id in the redis database returned an empty value")
	}
	ttl = client.TTL(ctx, fmt.Sprint(driverID)).Val()
	if ttl <= 0 {
		return "", 0, fmt.Errorf("the ttl the previous booking token is smaller than 0")
	}

	payload := make([]int, 3)
	err = sonic.UnmarshalString(client.Get(ctx, bookingID).Val(), &payload)
	if err != nil {
		return "", 0, err
	}
	partition := payload[0]

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
	client.Set(ctx, fmt.Sprint(driverID), id.String(), ttl)

	return token, 0, nil
}
