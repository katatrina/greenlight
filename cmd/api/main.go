package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Application version number
const version = "1.0.0"

// Configuration settings for our application.
type config struct {
	port int
	env  string
}

// application hold dependencies for our HTTP handlers, helpers, and middlewares.
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	// Read the value of the port and env command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new structured logger which writes log entries to the standard out stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := application{
		config: cfg,
		logger: logger,
	}

	// Declare a HTTP server which listens on the port provided in the config struct,
	// uses the servemux we created above as the handler, has some sensible timeout
	// settings and writes any log messages to the structured logger at Error level.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Start the HTTP server.
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
