// Generate random latitudes and longitudes for testing the stream
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var e env.Env

func init() {
	e.Load()

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
}

type coordinates struct {
	Timestamp int64   `json:"timestamp"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Heading   float64 `json:"heading"`
}

type response struct {
	DriverName            string        `json:"driver_name"`
	VehicleModel          string        `json:"vehicle_modal"`
	VehicleRegistrationNo string        `json:"vehicle_registration_no"`
	Coordinates           []coordinates `json:"cordinates"`
}

func main() {
	args := os.Args
	if len(args) != 3 {
		log.Debug().Msg("please provide the required environment and the booking ID")
		return
	}
	env := args[1]
	if env != "prd" && env != "stg" && env != "dev" {
		log.Debug().Msg("the provided environment is not valid")
		return
	}
	var url string
	switch env {
	case "prd":
		url = e.PrdURL
	case "stg":
		url = e.StgURL
	default:
		url = e.Domain
	}
	bookingID := args[2]

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/logs/view/%s", url, bookingID), nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to create the http request")
		return
	}
	req.Header.Add("Authorization", "Bearer "+e.AdminToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to send the request")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Debug().Int("status", resp.StatusCode).Msg("failed to generate the mock data")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read the response body")
		return
	}

	var payload response
	err = sonic.Unmarshal(body, &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal the response body")
		return
	}

	os.Remove(getPath(bookingID))
	os.Mkdir(getPath(bookingID), 0755)

	latFile, err := os.Create(getPathWBookingID(bookingID, "lat"))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the latitude file")
		return
	}
	defer latFile.Close()
	lonFile, err := os.Create(getPathWBookingID(bookingID, "lon"))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the longitude file")
		return
	}
	defer lonFile.Close()
	headingFile, err := os.Create(getPathWBookingID(bookingID, "headings"))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the headings file")
		return
	}
	defer headingFile.Close()

	for _, coordinate := range payload.Coordinates {
		_, err = fmt.Fprintf(latFile, "%f\n", coordinate.Lat)
		if err != nil {
			log.Fatal().AnErr("failed to write coordinate to the lat file", err)
			return
		}

		_, err = fmt.Fprintf(lonFile, "%f\n", coordinate.Lon)
		if err != nil {
			log.Fatal().AnErr("failed to write coordinate to the lon file", err)
			return
		}

		_, err = fmt.Fprintf(headingFile, "%f\n", coordinate.Heading)
		if err != nil {
			log.Fatal().AnErr("failed to write coordinate to the heading file", err)
			return
		}
	}
}

func getPath(fileName string) string {
	return "tests/stream/data/" + fileName
}

func getPathWBookingID(bookingID, fileName string) string {
	return getPath(bookingID) + "/" + fileName
}
