package jobs

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
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
	data, err := sonic.Marshal(payload{
		Items: []item{
			{
				Operation: "update",
				Key:       "secret",
				Value:     lib.GenerateToken(72),
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the payload")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://api.vercel.com/v1/edge-config/%s/items", e.EdgeConfig), bytes.NewBuffer(data))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the request")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Authorization", "Bearer "+e.VercelToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to send the reqeust to update the secret")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Err(fmt.Errorf("failed to update the secret")).Msg("failed to update the secret on the edge config")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "okay")
}
