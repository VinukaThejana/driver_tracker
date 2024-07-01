package bookings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	errs "github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"googlemaps.github.io/maps"
)

type bookingDetails struct {
	BookRefNo       *string
	BookPassengerNm *string
	BookingContact  *string
	BookPickUpAddr  *string
	BookDropAddr    *string
	DriverName      *string
	DriverContact   *string
	VehicleRegNo    *string
	VehicleModel    *string
	VehicleColor    *string
	BookTotal       *float64
}

type geo struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// view is a route that is used to get the booking information by the booking id
func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, errs.ErrBookingIDNotValid.Error())
		return
	}

	started := false
	val := c.R.DB.Get(r.Context(), bookingID).Val()
	if val != "" {
		started = true
	}
	var payload bookingDetails

	query := `SELECT
	bookings.BookRefNo,
	bookings.BookPassengerNm,
	bookings.BookingContact,
	bookings.BookPickUpAddr,
	bookings.BookDropAddr,
	bookings.BookTotal,
	bookings.DriverName,
	bookings.DriverContact,
	vehicles.VehicleRegNo,
	vehicles.VehicleModel,
	vehicles.VehicleColor
FROM
	Tbl_BookingDetails bookings
	INNER JOIN Tbl_VehicleDetails vehicles ON bookings.VehicleId = vehicles.VehicleId
WHERE
	bookings.BookRefNo = @BookRefNo`

	err := c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(
		&payload.BookRefNo,
		&payload.BookPassengerNm,
		&payload.BookingContact,
		&payload.BookPickUpAddr,
		&payload.BookDropAddr,
		&payload.BookTotal,
		&payload.DriverName,
		&payload.DriverContact,
		&payload.VehicleRegNo,
		&payload.VehicleModel,
		&payload.VehicleColor,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			lib.JSONResponse(w, http.StatusNotFound, errs.ErrBookingIDNotValid.Error())
			return
		}

		log.Error().Err(err).Msg("failed to get the query from the database")
		lib.JSONResponse(w, http.StatusInternalServerError, errs.ErrServer.Error())
		return
	}

	c.InitMap(e)

	BookPassengerNm := ""
	BookingContact := ""
	DriverName := ""
	DriverContact := ""
	VehicleRegNo := ""
	VehicleModel := ""
	VehicleColor := ""
	BookTotal := 0.0

	if payload.BookPassengerNm != nil {
		BookPassengerNm = *payload.BookPassengerNm
	}
	if payload.BookingContact != nil {
		BookingContact = *payload.BookingContact
	}
	if payload.BookTotal != nil {
		BookTotal = *payload.BookTotal
	}
	if payload.DriverName != nil {
		DriverName = *payload.DriverName
	}
	if payload.DriverContact != nil {
		DriverContact = *payload.DriverContact
	}
	if payload.VehicleRegNo != nil {
		VehicleRegNo = *payload.VehicleRegNo
	}
	if payload.VehicleModel != nil {
		VehicleModel = *payload.VehicleModel
	}
	if payload.VehicleColor != nil {
		VehicleColor = *payload.VehicleColor
	}

	pickups := []geo{}
	if payload.BookPickUpAddr != nil {
		payload, err := geocode(r.Context(), c, seperator(*payload.BookPickUpAddr))
		if err != nil {
			log.Error().Err(err).Msg("failed to geocode the pickup address")
			lib.JSONResponse(w, http.StatusInternalServerError, errs.ErrServer.Error())
			return
		}
		pickups = append(pickups, payload...)
	}

	dropoffs := []geo{}
	if payload.BookDropAddr != nil {
		payload, err := geocode(r.Context(), c, seperator(*payload.BookDropAddr))
		if err != nil {
			log.Error().Err(err).Msg("failed to geocode the drop address")
			lib.JSONResponse(w, http.StatusInternalServerError, errs.ErrServer.Error())
			return
		}

		dropoffs = append(dropoffs, payload...)
	}

	data := map[string]any{}
	data["active"] = started
	data["name"] = BookPassengerNm
	data["contact_no"] = BookingContact
	data["total"] = BookTotal
	data["driver_name"] = DriverName
	data["driver_contact"] = DriverContact
	data["vehicle_registration_no"] = VehicleRegNo
	data["vehicle_modal"] = VehicleModel
	data["vehicle_color"] = VehicleColor
	data["pickups"] = pickups
	data["dropoffs"] = dropoffs
	if started {
		data["stream"] = fmt.Sprintf("wss://%s/ws/stream/view/%s", e.Host, bookingID)
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
