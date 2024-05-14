// Package lib contains various library packages
package lib

import "github.com/VinukaThejana/go-utils/logger"

// LogFatal is a function that is used to Run various functions that need to crash if an error is found
func LogFatal(err error) {
	if err != nil {
		logger.Errorf(err)
	}
}
