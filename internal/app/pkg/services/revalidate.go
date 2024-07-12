package services

import (
	"fmt"
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
)

// Paths is a type that is used to revalidate the cache paths
type Paths string

const (
	// Dashboard is used to revalidate the dashboard
	Dashboard Paths = "dashboard"
	// History is used to revalidate the booking history
	History Paths = "history"
	// Alerts is used to revalidate the driver alerts
	Alerts Paths = "alerts"
)

// Revalidate is a service that is used to revalidate the path of a cache depending on the backend state
func Revalidate(e *env.Env, paths []Paths) {
	config := NewEdgeConfig(e.EdgeConfig, e.EdgeConfigReadToken, e.VercelToken)
	val, err := config.Read("secret")
	if err != nil {
		return
	}
	secret, ok := val.(string)
	if !ok {
		return
	}

	url := fmt.Sprintf("%s/api/admin/revalidate?secret=%s", e.DashboardURL, secret)
	for _, path := range paths {
		url = fmt.Sprintf("%s&paths=%s", url, path)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	client.Do(req)
}
