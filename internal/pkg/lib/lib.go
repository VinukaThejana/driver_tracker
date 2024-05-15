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

// ToStr is a function that is used to convert bad json strings to nicer json bytes
func ToStr(jsonStr string) ([]byte, error) {
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &payload)
	if err != nil {
		return []byte(jsonStr), err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return []byte(jsonStr), err
	}

	return payloadBytes, nil
}
