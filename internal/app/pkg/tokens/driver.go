package tokens

import (
	"fmt"
	"strconv"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/golang-jwt/jwt/v5"
)

// DriverToken is a struct that is used to perform actions that are related to the driver token
type DriverToken struct {
	C *connections.C
	E *env.Env
}

// Validate is a function that is used to validate the driver token
func (dt *DriverToken) Validate(str string) (isValid bool, token *jwt.Token) {
	token, err := jwt.Parse(str, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing algorithm was used")
		}

		return []byte(dt.E.DriverTokenSecret), nil
	})
	if err != nil || token == nil {
		return false, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil
	}

	role := claims["http://schemas.microsoft.com/ws/2008/06/identity/claims/role"].(string)
	if role != "D" {
		return false, nil
	}
	_, ok = claims["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/sid"].(string)
	if !ok {
		return false, nil
	}

	return true, token
}

// Get is a function that is used to get the claims of the given JWT token
func (dt *DriverToken) Get(token *jwt.Token) (driverID int, role string, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, "", fmt.Errorf("failed to map the claims of the jwt")
	}

	driverID, err = strconv.Atoi(claims["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/sid"].(string))
	if err != nil {
		return -1, "", fmt.Errorf("failed to validate the driverID")
	}

	role = claims["http://schemas.microsoft.com/ws/2008/06/identity/claims/role"].(string)
	return driverID, role, nil
}
