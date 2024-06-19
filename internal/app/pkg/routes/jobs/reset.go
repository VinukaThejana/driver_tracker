package jobs

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/errors"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

func reset(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	err := c.R.DB.FlushDB(r.Context()).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to flush the database")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	lib.JSONResponse(w, http.StatusOK, "cleared all the pending jobs ... ")
}
