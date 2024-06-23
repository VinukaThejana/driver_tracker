package bookings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/errors"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"googlemaps.github.io/maps"
)

type bookingDetails struct {
	BookRefNo      *string
	DriverName     *string
	ContactNo      *string
	VehicleModal   *string
	VehicleRegNo   *string
	BookPickUpAddr *string
	BookDropAddr   *string
}

type geo struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// view is a route that is used to get the booking information by the booking id
func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, "please provide a proper booking id")
		return
	}

	started := false
	val := c.R.DB.Get(r.Context(), bookingID).Val()
	if val != "" {
		started = true
	}
	var payload bookingDetails

	query := "SELECT BookRefNo, DriverName, ContactNo, VehicleModal, VehicleRegNo, BookPickUpAddr, BookDropAddr FROM Tbl_BookingDetails WHERE BookRefNo = @BookRefNo"
	err := c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(&payload.BookRefNo, &payload.DriverName, &payload.ContactNo, &payload.VehicleModal, &payload.VehicleRegNo, &payload.BookPickUpAddr, &payload.BookDropAddr)
	if err != nil {
		log.Error().Err(err).Msg("failed to get the query from the database")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	c.InitMap(e)

	DriverName := ""
	ContactNo := ""
	VehicleModal := ""
	VehicleRegNo := ""

	pickups := []geo{}
	dropoffs := []geo{}

	if payload.DriverName != nil {
		DriverName = *payload.DriverName
	}
	if payload.ContactNo != nil {
		ContactNo = *payload.ContactNo
	}
	if payload.VehicleModal != nil {
		VehicleModal = *payload.VehicleModal
	}
	if payload.VehicleRegNo != nil {
		VehicleRegNo = *payload.VehicleRegNo
	}
	if payload.BookPickUpAddr != nil {
		payload, err := geocode(r.Context(), c, seperator(*payload.BookPickUpAddr))
		if err != nil {
			log.Error().Err(err).Msg("failed to geocode the pickup address")
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}
		pickups = append(pickups, payload...)
	}
	if payload.BookDropAddr != nil {
		payload, err := geocode(r.Context(), c, seperator(*payload.BookDropAddr))
		if err != nil {
			log.Error().Err(err).Msg("failed to geocode the drop address")
			lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
			return
		}

		dropoffs = append(dropoffs, payload...)
	}

	data := map[string]any{}
	data["active"] = started
	data["driver_name"] = DriverName
	data["contact_no"] = ContactNo
	data["vehicle_modal"] = VehicleModal
	data["vehicle_registration_no"] = VehicleRegNo
	data["pickups"] = pickups
	data["dropoffs"] = dropoffs
	if started {
		data["stream"] = fmt.Sprintf("ws://%s/ws/stream/view/%s", e.Host, bookingID)
	}

	lib.JSONResponseWInterface(w, http.StatusOK, data)
}

func seperator(str string) []string {
	if strings.Contains(str, "|") {
		return strings.Split(str, "|")
	}

	return []string{str}
}

func geocode(ctx context.Context, c *connections.C, data []string) ([]geo, error) {
	pickups := []geo{}

	for _, pickup := range data {
		route, err := c.M.Geocode(ctx, &maps.GeocodingRequest{
			Address: pickup,
		})
		if err != nil {
			return []geo{}, nil
		}
		if len(route) == 0 {
			return []geo{}, fmt.Errorf("failed to geocode the street address")
		}

		pickups = append(pickups, geo{
			Lat: route[0].Geometry.Location.Lat,
			Lon: route[0].Geometry.Location.Lng,
		})
	}

	return pickups, nil
}
