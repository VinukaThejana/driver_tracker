package connections

import (
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"googlemaps.github.io/maps"
)

// InitMap is a function that is used to initialize the Google maps connections
func (c *C) InitMap(e *env.Env) {
	client, err := maps.NewClient(maps.WithAPIKey(e.GoogleMapsAPIKey))
	if err != nil {
		lib.LogFatal(err)
	}

	c.M = client
}
