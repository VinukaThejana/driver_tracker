package bookings

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type bookingDetails struct {
	BookRefNo    *string
	DriverName   *string
	ContactNo    *string
	VehicleModal *string
	VehicleRegNo *string
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
	fmt.Printf("started: %v\n", started)

	var payload bookingDetails

	query := "SELECT BookRefNo, DriverName, ContactNo, VehicleModal, VehicleRegNo FROM Tbl_BookingDetails WHERE BookRefNo = @BookRefNo"
	err := c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(&payload.BookRefNo, &payload.DriverName, &payload.ContactNo, &payload.VehicleModal, &payload.VehicleRegNo)
	if err != nil {
		log.Error().Err(err).Msg("failed to get the query from the database")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong, please try again later")
		return
	}

	DriverName := ""
	ContactNo := ""
	VehicleModal := ""
	VehicleRegNo := ""
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

	data := map[string]interface{}{}
	data["active"] = started
	data["driver_name"] = DriverName
	data["contact_no"] = ContactNo
	data["vehicle_modal"] = VehicleModal
	data["vehicle_registration_no"] = VehicleRegNo
	if started {
		data["stream"] = fmt.Sprintf("ws://%s/ws/stream/view/%s", e.Host, bookingID)
	}

	lib.JSONResponseWInterface(w, http.StatusOK, data)
}
