// Package lib contains app specific libraries
package lib

import (
	"fmt"
)

// L is used to get the last location from redis
func L(partition any) string {
	_, ok := partition.(int)
	if ok {
		return fmt.Sprintf("l%d", partition)
	}
	return fmt.Sprintf("l%s", partition)
}

// C is used to get the maximum number of connections in redis
func C(partition any) string {
	_, ok := partition.(int)
	if ok {
		return fmt.Sprintf("c%d", partition)
	}
	return fmt.Sprintf("c%s", partition)
}

// N is a key used to get information regarding the booking from redis
func N(partition any) string {
	_, ok := partition.(int)
	if ok {
		return fmt.Sprintf("n%d", partition)
	}
	return fmt.Sprintf("n%s", partition)
}
