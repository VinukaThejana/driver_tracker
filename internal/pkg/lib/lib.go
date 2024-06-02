// Package lib contains various library packages
package lib

import (
	"github.com/VinukaThejana/go-utils/logger"
	"github.com/bytedance/sonic"
)

// LogFatal is a function that is used to Run various functions that need to crash if an error is found
func LogFatal(err error) {
	if err != nil {
		logger.Errorf(err)
	}
}

// ToStr is a function that is used to convert bad json strings to nicer json bytes
func ToStr(jsonStr string) ([]byte, error) {
	var (
		data    map[string]interface{}
		payload []byte
		err     error
	)

	if err = sonic.UnmarshalString(jsonStr, &data); err != nil {
		return nil, err
	}
	if payload, err = sonic.Marshal(data); err != nil {
		return nil, err
	}

	return payload, nil
}
