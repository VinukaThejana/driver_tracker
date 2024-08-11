package stream

import (
	"fmt"
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/enums"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
)

func create(w http.ResponseWriter, r *http.Request, e *env.Env, _ *connections.C) {
	type res struct {
		WebsocketURL string `json:"websocketURL"`
	}

	bookingTokenID := r.Context().Value(middlewares.BookingTokenID).(string)
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partition := r.Context().Value(middlewares.PartitionNo).(int)

	response := res{}
	if e.Env == string(enums.Dev) {
		response.WebsocketURL = fmt.Sprintf("ws://%s/ws/stream/add/%s/%d/%d", e.WebsocketURL, bookingTokenID, driverID, partition)
	} else {
		response.WebsocketURL = fmt.Sprintf("wss://%s/ws/stream/add/%s/%d/%d", e.WebsocketURL, bookingTokenID, driverID, partition)
	}

	lib.JSONResponseWInterface(w, http.StatusOK, response)
}
