package tokens

import (
	"fmt"
	"strconv"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/golang-jwt/jwt/v5"
)

// AdminToken is a struct that is used to perform actions that are related to the admin token
type AdminToken struct {
	C *connections.C
	E *env.Env
}

// Validate is a function that is used to validate the admin token
func (at *AdminToken) Validate(str string) (isValid bool, token *jwt.Token) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing algorithm was used")
		}

		return []byte(at.E.AdminTokenSecret), nil
	})
	if err != nil || token == nil {
		return false, nil
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil
	}

	return true, token
}

// Get is a function that is used to get the claims of the given JWT token
func (at *AdminToken) Get(token *jwt.Token) (id int, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, fmt.Errorf("failed to map the claims of the jwt")
	}

	val, ok := claims["sub"].(string)
	if !ok {
		return -1, fmt.Errorf("cannot parse the driver id")
	}
	id, err = strconv.Atoi(val)
	if err != nil {
		return -1, err
	}

	return id, nil
}
