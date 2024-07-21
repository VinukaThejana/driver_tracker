package lib

// JobStatus is used to classify the job status codes
type JobStatus int

const (
	// NotAccepted is when the job is still not accepted by the driver
	NotAccepted JobStatus = 0
	// Accepted is when the the job is accepted by the driver
	Accepted JobStatus = 1
	// OnTheWay is when the driver is on the way to the pickup point
	OnTheWay JobStatus = 2
	// PickupPoint is when the driver is in the pickup point
	PickupPoint JobStatus = 3
	// PassengerOnBoard is when the passenger is onboard
	PassengerOnBoard JobStatus = 4
	// Clear is when the job is cleared by the driver
	Clear JobStatus = 5
	// DefaultStatus is the default status when the status is not provided
	DefaultStatus JobStatus = OnTheWay
)
