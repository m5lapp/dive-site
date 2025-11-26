package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *app) serve() error {
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
	}

	srv := &http.Server{
		Addr:         app.config.addr,
		ErrorLog:     slog.NewLogLogger(app.log.Handler(), slog.LevelError),
		Handler:      app.routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  5 * time.Second,
		TLSConfig:    tlsConfig,
		WriteTimeout: 10 * time.Second,
	}

	var err error
	shutdownError := make(chan error)

	go func() {
		// Catch SIGINT and SIGTERM system calls and handle them gracefully.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.log.Info("Caught OS signal, shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(
			context.Background(),
			app.config.termPeriod,
		)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.log.Info("Starting server", "addr", app.config.addr)

	if app.config.tlsCert != "" && app.config.tlsKey != "" {
		err = srv.ListenAndServeTLS(app.config.tlsCert, app.config.tlsKey)
	} else {
		app.log.Warn("No TLS cert and key was provided, insecure HTTP will be used")
		err = srv.ListenAndServe()
	}

	// An http.ErrServerClosed error is expected when the server is shutting
	// down properly. If we receive anything else then something has gone wrong.
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Check for any errors in srv.Shutdown() process.
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.log.Info("Server stopped successfully.")
	return nil
}
