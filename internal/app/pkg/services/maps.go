package services

import (
	"context"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"googlemaps.github.io/maps"
)

// Geo lat-lon direction mapping
type Geo struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Geocode convert the string paths to lat-lon paths
func Geocode(
	ctx context.Context,
	e *env.Env,
	c *connections.C,
	init bool,
	paths []string,
) []Geo {
	if init {
		c.InitMap(e)
	}

	payload := []Geo{}

	for _, path := range paths {
		route, err := c.M.Geocode(ctx, &maps.GeocodingRequest{
			Address: path,
		})
		if err != nil {
			return []Geo{}
		}
		if len(route) == 0 {
			continue
		}

		payload = append(payload, Geo{
			Lat: route[0].Geometry.Location.Lat,
			Lon: route[0].Geometry.Location.Lng,
		})
	}

	return payload
}
