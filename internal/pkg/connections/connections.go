// Package connections contains connections to various third party services
package connections

// C contains all third pary connections
type C struct {
	// R contains all Redis related databases
	R *Redis
	// K contains all Kafka writers
	K *KafkaWriters
}

// Close is a function that is used to close all the connections
func (c *C) Close() {
	// close all the kafka writers
	c.K.B.Close()

	// close all the redis clients
	c.R.DB.Close()
}
