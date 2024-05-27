// Spoton Cars streaming platfrom
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/routes"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	e         env.Env
	connector connections.C

	addR    routes.Route
	viewR   routes.Route
	healthR routes.Route
	createR routes.Route
)

func init() {
	e.Load()
	connector.InitRedis(&e)

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	addR = &routes.Add{
		E: &e,
		C: &connector,
	}
	viewR = &routes.View{
		E: &e,
		C: &connector,
	}
	healthR = &routes.Health{
		E: &e,
		C: &connector,
	}
	createR = &routes.CreateStream{
		E: &e,
		C: &connector,
	}
}

func router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.RequestID)
	r.Use(middlewares.RealIP)
	r.Use(middlewares.Logger)
	r.Use(middlewares.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"Content-Type", "X-CSRF-Token"},
	}))

	r.MethodFunc(addR.Method(), addR.Path(), addR.Handler)
	r.MethodFunc(healthR.Method(), healthR.Path(), healthR.Handler)

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.CheckContentTypeIsJSON)
		r.MethodFunc(viewR.Method(), viewR.Path(), viewR.Handler)
		r.MethodFunc(createR.Method(), createR.Path(), createR.Handler)
	})

	return r
}

func shutdown(ctx context.Context, server *http.Server) {
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("failed to shudown server gracefully")
		return
	}

	log.Info().Msg("server shutdown gracefully")
}

func main() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", e.Port),
		Handler: router(),
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info().Msgf("started the HTTP server and listening on port : %d", e.Port)
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("HTTP server error")
				return
			}
		}
	}()

	select {
	case sig := <-signalCh:
		log.Info().Str("cause", "signal").Str("signal", sig.String()).Msg("shutting down server")
		shutdown(ctx, server)
		cancel()
	case <-ctx.Done():
		log.Info().Msg("context cancelled, shutting down the server")
		shutdown(ctx, server)
		cancel()
	}
}
