// Package lib contains various library packages
package lib

import (
	"encoding/json"

	"github.com/VinukaThejana/go-utils/logger"
)

// LogFatal is a function that is used to Run various functions that need to crash if an error is found
func LogFatal(err error) {
	if err != nil {
		logger.Errorf(err)
	}
}
<<<<<<< HEAD
<<<<<<< Updated upstream
=======
=======
>>>>>>> 8a5c40b (feat(added-lib): Added a package to convery string to bytes)

// ToStr is a function that is used to convert bad json strings to nicer json bytes
func ToStr(jsonStr string) ([]byte, error) {
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &payload)
	if err != nil {
<<<<<<< HEAD
		return []byte(jsonStr), err
=======
		return nil, err
>>>>>>> 8a5c40b (feat(added-lib): Added a package to convery string to bytes)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
<<<<<<< HEAD
		return []byte(jsonStr), err
=======
		return nil, err
>>>>>>> 8a5c40b (feat(added-lib): Added a package to convery string to bytes)
	}

	return payloadBytes, nil
}
<<<<<<< HEAD
>>>>>>> Stashed changes
=======
>>>>>>> 8a5c40b (feat(added-lib): Added a package to convery string to bytes)
