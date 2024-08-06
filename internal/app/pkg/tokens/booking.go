package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// BookingToken is a token that is used to identify the driver that is posting to the booking_id
type BookingToken struct {
	C *connections.C
	E *env.Env
}

// NewBookingToken is a function that is used to create a new booking token instance
func NewBookingToken(e *env.Env, c *connections.C) *BookingToken {
	return &BookingToken{
		C: c,
		E: e,
	}
}

// Createtoken is a function that is used to only create the booking token without manipulating the redis state
func (bt *BookingToken) Createtoken(
	driverID,
	partitionNo int,
	bookingID string,
	duration time.Duration,
) (id uuid.UUID, token string, err error) {
	now := time.Now().UTC()

	id, err = uuid.NewUUID()
	if err != nil {
		return uuid.UUID{}, "", err
	}

	claims := make(jwt.MapClaims)

	claims["sub"] = id.String()
	claims["exp"] = now.Add(duration).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["driver_id"] = driverID
	claims["booking_id"] = bookingID
	claims["partition_no"] = partitionNo

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(bt.E.BookingTokenSecret))
	if err != nil {
		return uuid.UUID{}, "", err
	}

	return id, token, nil
}

// Create is a function that is used to create the booking token
func (bt *BookingToken) Create(
	ctx context.Context,
	driverID int,
	bookingID string,
	partitionNo int,
	newOffset int,
	payload string,
	pickup services.Geo,
) (token string, err error) {
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", bt.E.BookingTokenExpires))
	if err != nil {
		return "", err
	}

	nPayload, err := sonic.MarshalString(_lib.SetN(bookingID, newOffset))
	if err != nil {
		return "", err
	}

	id, token, err := bt.Createtoken(driverID, partitionNo, bookingID, duration)
	if err != nil {
		return "", err
	}

	pickupStr, err := sonic.MarshalString(map[string]any{
		"lat":       pickup.Lat,
		"lon":       pickup.Lon,
		"heading":   0,
		"accuracy":  -1,
		"status":    _lib.DefaultStatus,
		"timestamp": time.Now().UTC().Unix(),
	})
	if err != nil {
		return "", err
	}
	driverDetails, err := sonic.MarshalString(_lib.SetDriverID(id.String(), bookingID, partitionNo))
	if err != nil {
		return "", err
	}

	pipe := bt.C.R.DB.Pipeline()
	pipe.SetNX(ctx, fmt.Sprint(driverID), driverDetails, duration)
	pipe.SetNX(ctx, bookingID, payload, duration)
	pipe.SetNX(ctx, _lib.L(partitionNo), pickupStr, duration)
	pipe.SetNX(ctx, _lib.C(partitionNo), 0, duration)
	pipe.SetNX(ctx, _lib.N(partitionNo), nPayload, duration+12*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Validate is a function that is used to validate the booking token
func (bt *BookingToken) Validate(
	ctx context.Context,
	str string,
	validateRemote bool,
) (isValid bool, token *jwt.Token) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing algorithm was used")
		}

		return []byte(bt.E.BookingTokenSecret), nil
	})
	if err != nil || token == nil {
		return false, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return false, nil
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		return false, nil
	}

	driverID, ok := claims["driver_id"].(float64)
	if !ok {
		return false, nil
	}

	if validateRemote {
		client := bt.C.R.DB

		val := client.Get(ctx, fmt.Sprint(int(driverID))).Val()
		if val == "" {
			return false, nil
		}
		DriverID := _lib.NewDriverID()
		err = sonic.UnmarshalString(val, &DriverID)
		if err != nil {
			return false, nil
		}
		if DriverID[_lib.DriverIDDriverToken] != id.String() {
			return false, nil
		}
	}

	return true, token
}

// Get is a function that is used to get the details from the booking token
func (bt *BookingToken) Get(
	token *jwt.Token,
) (id string, driverID int, bookingID string, partitionNo int, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", 0, "", 0, fmt.Errorf("failed to map the token to claims")
	}

	id = claims["sub"].(string)
	driverID = int(claims["driver_id"].(float64))
	bookingID = claims["booking_id"].(string)
	partitionNo = int(claims["partition_no"].(float64))

	return id, driverID, bookingID, partitionNo, nil
}
