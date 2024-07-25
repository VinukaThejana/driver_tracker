package lib

// RedisBookingID is used to represent the booking in Redis
type RedisBookingID int

// [
//  partition_no,
//  kafka_last_offset,
//  driver_id
// ]

const (
	// BookingIDPartitionNo is used to represent the partition number under the booking ID
	BookingIDPartitionNo RedisBookingID = iota
	// BookingIDLastOffset is used to get the lastoffset of kafka when the booking is created
	BookingIDLastOffset
	// BookingIDDriverID is used to get the BookingIDDriverID of the driver in the active booking ID
	BookingIDDriverID
)

// RedisN is used to represent the backup that is stored regarding the booking
// this is used to get information regarding the booking if the driver/admin fails to
// clear the stream
type RedisN int

// [
//  booking_id,
//  kafka_last_offset
// ]

const (
	// NBookingID is used to get booking ID from the Redis backup
	NBookingID RedisN = iota
	// NLastOffset is used to get the last offset from kafka in the event of booking creation
	NLastOffset
)

// RedisDriverID is used to get the driver booking token and other driver related items from the redis
// database for authenticating and performing driver related actions
type RedisDriverID int

// [
//  driver_token_id,
//  booking_id,
//  partition_no
// ]

const (
	// DriverIDDriverToken is used to get the driver token for validating the driver
	DriverIDDriverToken RedisDriverID = iota
	// DriverIDBookingID is used to get the active booking ID of the driver
	DriverIDBookingID
	// DriverIDPartitionNo is used to get the partition number of the given driver
	DriverIDPartitionNo
)
