package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
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

// Create is a function that is used to create the booking token
func (bt *BookingToken) Create(
	ctx context.Context,
	driverID int,
	bookingID string,
	partitionNo int,
	newOffset int,
	payload string,
) (token string, err error) {
	now := time.Now().UTC()

	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", bt.E.BookingTokenExpires))
	if err != nil {
		return "", err
	}

	nPayload, err := sonic.MarshalString([]string{bookingID, fmt.Sprint(newOffset)})
	if err != nil {
		return "", err
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
		return "", err
	}

	pipe := bt.C.R.DB.Pipeline()
	pipe.SetNX(ctx, id.String(), true, duration)
	pipe.SetNX(ctx, fmt.Sprint(partitionNo), id.String(), duration)
	pipe.SetNX(ctx, bookingID, payload, duration)
	pipe.SetNX(ctx, fmt.Sprintf("n%d", partitionNo), nPayload, duration)
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

	id, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return false, nil
	}

	client := bt.C.R.DB

	val := client.Get(ctx, id.String()).Val()
	if val == "" {
		return false, nil
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
