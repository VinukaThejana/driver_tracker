package jobs

import (
	"fmt"
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

type item struct {
	Operation string `json:"operation"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

type payload struct {
	Items []item `json:"items"`
}

func rotate(w http.ResponseWriter, _ *http.Request, e *env.Env, _ *connections.C) {
	config := services.NewEdgeConfig(e.EdgeConfig, e.EdgeConfigReadToken, e.VercelToken)
	err := config.Update([]services.EdgeConfigItem{
		{
			Operation: services.EdgeConfigUpdate,
			Key:       "secret",
			Value:     lib.GenerateToken(72),
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to generate the secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "okay")
}
