package tokens

import (
	"fmt"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
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

		return []byte(""), nil
	})
	if err != nil || token == nil {
		return false, nil
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil
	}

	return true, nil
}
