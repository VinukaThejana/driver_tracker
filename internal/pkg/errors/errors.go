// Package errors contains essential errors
package errors

import "fmt"

var (
	// ErrServer is to indicate an internal server error
	ErrServer = fmt.Errorf("something went wrong, please try again later")
	// ErrBusy is to indicate that the server is busy at the moment
	ErrBusy = fmt.Errorf("server is busy right now, please try again later")
	// ErrUnauthorized is to indicate the user is unauthorized to perform the given operation
	ErrUnauthorized = fmt.Errorf("you are not authorized to perform this operation")
	// ErrBadRequest is to indicate that the request is a bad request
	ErrBadRequest = fmt.Errorf("invalid data, please check and try agan")
	// ErrNotAdmin is to indicate that the requesting user is not the admin
	ErrNotAdmin = fmt.Errorf("you are not an admin to perform this operation")
	// ErrBookingIDNotValid is to indicate that the given booking id is not valid
	ErrBookingIDNotValid = fmt.Errorf("booking id you provided is not valid")
	// ErrUnsuportedMedia is to indicate that the request body that the client is providing is not supported
	ErrUnsuportedMedia = fmt.Errorf("request body is not supported")
)
