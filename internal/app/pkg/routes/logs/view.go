package logs

import (
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	_errors "github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type bookingDetails struct {
	DriverName     *string
	VehicleModal   *string
	VehicleRegNo   *string
	BookPickUpAddr *string
	BookDropAddr   *string
}

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, _errors.ErrBookingIDNotValid.Error())
	}

	err := c.InitStorage(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize the storage client")
		lib.JSONResponse(w, http.StatusInternalServerError, _errors.ErrServer.Error())
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	reader, err := object.NewReader(r.Context())
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"booking_id : %s\tfailed to initialize the reader, either because the booking id is not valid or some autentication error",
				bookingID,
			)
		lib.JSONResponse(w, http.StatusBadRequest, _errors.ErrBookingIDNotValid.Error())
		return
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("failed to read the data")
		lib.JSONResponse(w, http.StatusInternalServerError, _errors.ErrServer.Error())
		return
	}
	if data == nil {
		lib.JSONResponse(w, http.StatusBadRequest, _errors.ErrBookingIDNotValid.Error())
		return
	}

	var payload any
	err = sonic.Unmarshal(data, &payload)
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"payload : %v\tfailed to marshal the data from the payload",
				payload,
			)
		lib.JSONResponse(w, http.StatusInternalServerError, _errors.ErrServer.Error())
		return
	}

	var bd bookingDetails

	query := `
SELECT
	DriverName,
	VehicleModal,
	VehicleRegNo,
	BookPickUpAddr,
	BookDropAddr
FROM
	Tbl_BookingDetails
WHERE
  BookRefNo = @BookRefNo
`

	err = c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(
		&bd.DriverName,
		&bd.VehicleModal,
		&bd.VehicleRegNo,
		&bd.BookPickUpAddr,
		&bd.BookDropAddr,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("failed to get the query from the database")
			lib.JSONResponse(w, http.StatusInternalServerError, _errors.ErrServer.Error())
			return
		}
	}
	DriverName := ""
	VehicleModal := ""
	VehicleRegNo := ""

	if bd.DriverName != nil {
		DriverName = *bd.DriverName
	}
	if bd.VehicleModal != nil {
		VehicleModal = *bd.VehicleModal
	}
	if bd.VehicleRegNo != nil {
		VehicleRegNo = *bd.VehicleRegNo
	}

	pickups := []services.Geo{}
	if bd.BookPickUpAddr != nil {
		pickups = append(pickups, services.Geocode(r.Context(), e, c, true, lib.Seperator(*bd.BookPickUpAddr, "|"))...)
	}

	dropoffs := []services.Geo{}
	if bd.BookDropAddr != nil {
		dropoffs = append(dropoffs, services.Geocode(r.Context(), e, c, true, lib.Seperator(*bd.BookDropAddr, "|"))...)
	}

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"driver_name":             DriverName,
		"vehicle_modal":           VehicleModal,
		"vehicle_registration_no": VehicleRegNo,
		"cordinates":              payload,
		"pickups":                 pickups,
		"dropoffs":                dropoffs,
	})
}
