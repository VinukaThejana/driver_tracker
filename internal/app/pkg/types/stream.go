package types

import (
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
)

// LocationUpdate represents a single location update in a stream.
// It contains geographical coordinates, accuracy, heading, and status information.
// The Lat and Lon fields are required and validated as latitude and longitude respectively.
// Status, if provided, must be one of the values 0, 1, 2, 3, 4, or 5.
type LocationUpdate struct {
	Accuracy      *float64 `json:"accuracy"`
	Status        *int64   `json:"status" validate:"omitempty,oneof=0 1 2 3 4 5"`
	Heading       *float64 `json:"heading"`
	LocationIndex *int     `json:"location_index"`
	Lat           float64  `json:"lat" validate:"required,latitude"`
	Lon           float64  `json:"lon" validate:"required,longitude"`
}

// GetBlob represents a function that is used replace the the LocationUpdate with
// a map by adding default values to the fields that are not presented
func (location *LocationUpdate) GetBlob() map[string]any {
	return map[string]any{
		"lat": location.Lat,
		"lon": location.Lon,
		"heading": func() float64 {
			if location.Heading == nil {
				return 0
			}
			return *location.Heading
		}(),
		"accuracy": func() float64 {
			if location.Accuracy == nil {
				return -1
			}
			return *location.Accuracy
		}(),
		"location_index": func() int {
			if location.LocationIndex == nil {
				return 0
			}
			return *location.LocationIndex
		}(),
		"status": func() int {
			if location.Status == nil {
				return int(_lib.DefaultStatus)
			}

			return int(*location.Status)
		}(),
	}
}
