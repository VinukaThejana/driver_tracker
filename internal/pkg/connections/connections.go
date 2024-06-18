// Package connections contains connections to various third party services
package connections

import (
	"database/sql"

	"googlemaps.github.io/maps"
)

// C contains all third pary connections
type C struct {
	// R contains all Redis related databases
	R *Redis
	// K contains all Kafka writers
	K *KafkaWriters
	// DB contains the Database connection
	DB *sql.DB
	// M contains the connection to Google maps
	M *maps.Client
}

// Close is a function that is used to close all the connections
func (c *C) Close() {
	// close all the kafka writers
	c.K.B.Close()

	// close all the redis clients
	c.R.DB.Close()

	// close the connection to the database
	c.DB.Close()
}
