package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
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
	driverID string,
	bookingID string,
	partitionNo int,
) (token string, err error) {
	now := time.Now().UTC()

	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	claims := make(jwt.MapClaims)
	claims["sub"] = id.String()
	claims["exp"] = now.Add(bt.E.BookingTokenExpires).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["driver_id"] = driverID
	claims["booking_id"] = bookingID
	claims["partition_no"] = partitionNo

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(bt.E.BookingTokenSecret))
	if err != nil {
		return "", err
	}

	client := bt.C.R.DB
	err = client.SetNX(ctx, id.String(), true, bt.E.BookingTokenExpires).Err()
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
) (id, driverID, bookingID string, partitionNo int, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", "", 0, err
	}

	id = claims["sub"].(string)
	driverID = claims["driver_id"].(string)
	bookingID = claims["booking_id"].(string)
	partitionNo = claims["partition_no"].(int)

	return id, driverID, bookingID, partitionNo, nil
}
