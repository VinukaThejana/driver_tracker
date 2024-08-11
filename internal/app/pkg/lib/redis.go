package lib

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

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

	// BookingIDSize is used to get the booking ID size in Redis
	BookingIDSize
)

// NewBookingID is a function that is used to create a new booking ID array
func NewBookingID() [BookingIDSize]int {
	return [BookingIDSize]int{}
}

// SetBookingID is a function that is used to create a fully populated BookingID array
func SetBookingID(
	partitionNo,
	lastOffset,
	driverID int,
) [BookingIDSize]int {
	BookingID := NewBookingID()

	BookingID[BookingIDPartitionNo] = partitionNo
	BookingID[BookingIDLastOffset] = lastOffset
	BookingID[BookingIDDriverID] = driverID

	return BookingID
}

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

	// NSize is used to get the size of the n partition in Redis
	NSize
)

// NewN is a function that is used to create a new N partition
func NewN() [NSize]string {
	return [NSize]string{}
}

// SetN is used to set the N partition
func SetN(
	bookingID string,
	lastOffset int,
) [NSize]string {
	N := NewN()

	N[NBookingID] = bookingID
	N[NLastOffset] = fmt.Sprint(lastOffset)

	return N
}

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

	// DriverIDSize is used to get the size of the driver ID array
	DriverIDSize
)

// NewDriverID is used to create a new driver ID array
func NewDriverID() [DriverIDSize]string {
	return [DriverIDSize]string{}
}

// SetDriverID is a function that is used to create the DriverID array
func SetDriverID(
	driverToken,
	bookingID string,
	partitionNo int,
) [DriverIDSize]string {
	DriverID := NewDriverID()

	DriverID[DriverIDDriverToken] = driverToken
	DriverID[DriverIDBookingID] = bookingID
	DriverID[DriverIDPartitionNo] = fmt.Sprint(partitionNo)

	return DriverID
}

// DelBooking is a function that is used to remove a booking and all its related components
// from the redis database
func DelBooking(
	ctx context.Context,
	client *redis.Client,
	driverID int,
	bookingID string,
	partition int,
) error {
	pipe := client.Pipeline()

	pipe.Del(ctx, fmt.Sprint(driverID))
	pipe.Del(ctx, bookingID)
	pipe.Del(ctx, L(partition))
	pipe.Del(ctx, C(partition))
	pipe.Del(ctx, N(partition))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Free is used to deallocate the used partition for upcomming jobs
func Free(ctx context.Context, client *redis.Client, key string, partition int) {
	pipe := client.Pipeline()
	pipe.Del(ctx, L(partition))
	pipe.SRem(ctx, key, partition)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"partition : %d\tfailed to remove the partition",
				partition,
			)
	}
}
