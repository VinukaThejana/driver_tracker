package logs

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type bookingDetails struct {
	DriverName   *string
	VehicleModal *string
	VehicleRegNo *string
}

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
	}

	err := c.InitStorage(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize the storage client")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	reader, err := object.NewReader(r.Context())
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID).Msg("failed to initialize the reader, either because the booking id is not valid or some autentication error")
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("failed to read the data")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	if data == nil {
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	var payload any
	err = sonic.Unmarshal(data, &payload)
	if err != nil {
		log.Error().Err(err).Interface("payload", payload).Msg("failed to marshal the data from the payload")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	var bookingDetails bookingDetails
	query := "SELECT DriverName, VehicleModal, VehicleRegNo FROM Tbl_BookingDetails WHERE BookRefNo = @BookRefNo"
	err = c.DB.QueryRow(query, sql.Named("BookRefNo", bookingID)).Scan(&bookingDetails.DriverName, &bookingDetails.VehicleModal, &bookingDetails.VehicleRegNo)
	if err != nil {
		log.Error().Err(err).Msg("failed to get the query from the database")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	DriverName := ""
	VehicleModal := ""
	VehicleRegNo := ""

	if bookingDetails.DriverName != nil {
		DriverName = *bookingDetails.DriverName
	}
	if bookingDetails.VehicleModal != nil {
		VehicleModal = *bookingDetails.VehicleModal
	}
	if bookingDetails.VehicleRegNo != nil {
		VehicleRegNo = *bookingDetails.VehicleRegNo
	}

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"driver_name":             DriverName,
		"vehicle_modal":           VehicleModal,
		"vehicle_registration_no": VehicleRegNo,
		"cordinates":              payload,
	})
}
