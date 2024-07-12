package jobs

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

func reset(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	err := c.R.DB.FlushDB(r.Context()).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to flush the database")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	services.Revalidate(e, []services.Paths{
		services.Dashboard,
	})

	lib.JSONResponse(w, http.StatusOK, "cleared all the pending jobs ... ")
}
