// Package errors contains essential errors
package errors

import "fmt"

var (
	ErrServer            = fmt.Errorf("something went wrong, please try again later")
	ErrBusy              = fmt.Errorf("server is busy right now, please try again later")
	ErrUnauthorized      = fmt.Errorf("you are not authorized to perform this operation")
	ErrBadRequest        = fmt.Errorf("invalid data, please check and try agan")
	ErrNotAdmin          = fmt.Errorf("you are not an admin to perform this operation")
	ErrBookingIDNotValid = fmt.Errorf("booking id you provided is not valid")
)
