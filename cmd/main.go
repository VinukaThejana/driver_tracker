// Spoton Cars streaming platfrom
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/denisenkom/go-mssqldb/azuread"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/routes"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/websockets"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/enums"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	e         env.Env
	connector connections.C

	rt routes.Route
	ws websockets.Websocket

	viewW websockets.Websocket
	addW  websockets.Websocket
)

func init() {
	e.Load()

	if e.Env == string(enums.Dev) {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out: os.Stderr,
		})
	}

	// NOTE: Initialize the most essential services only
	// No doing so will effect the startup time of the container
	connector.InitRedis(&e)
	connector.InitKafkaWriters(&e)
	connector.InitDB(&e)

	rt = routes.Route{
		E: &e,
		C: &connector,
	}
	ws = websockets.Websocket{
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
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Mount("/", rt.Routes())
	r.Mount("/ws", ws.Websocket())

	return r
}

func shutdown(ctx context.Context, engine *nbhttp.Engine) {
	shutdownCtx, shutdownCtxCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCtxCancel()

	if err := engine.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("failed to shudown server gracefully")
		return
	}

	log.Info().Msg("server shutdown gracefully")
}

func main() {
	defer connector.Close()

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(e.Port),
		Handler: router(),
	}

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:       "tcp",
		Addrs:         []string{server.Addr},
		KeepaliveTime: time.Second * 60 * 60,
		Handler:       server.Handler,
	})

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info().
		Msgf("started the HTTP server and listening on port : %d", e.Port)
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			if err := engine.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("HTTP server error")
				return
			}
		}
	}()

	select {
	case sig := <-signalCh:
		log.Info().
			Str("cause", "signal").
			Str("signal", sig.String()).
			Msg("shutting down server")
		shutdown(ctx, engine)
		break
	case <-ctx.Done():
		log.Info().Msg("context cancelled, shutting down the server")
		shutdown(ctx, engine)
		break
	}
}
