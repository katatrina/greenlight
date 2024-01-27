package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"github.com/katatrina/greenlight/util"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Declare a string containing the application version number.
const version = "1.0.0"

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middlewares.
type application struct {
	logger *slog.Logger
	config util.Config
	store  *db.Store
}

func main() {
	config, err := util.LoadConfig("./app_config.yaml")

	// Read the value of the port and env command-line flags into the config struct.
	// Default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&config.ServerPort, "port", 4000, "API server port")
	flag.StringVar(&config.Environment, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialized a new structured logger which writes log entries to the standard out stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Initialize a new connection pool to our database.
	connPool, err := openDB(config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database connection pool established")

	store := db.NewStore(connPool)

	app := &application{
		logger,
		config,
		store,
	}
	// Declare an HTTP server which listens to the port provided in the config struct,
	// uses the router returned by routes method, as some sensible timeout settings and
	// writes any log messages to the structured logger at Error level.
	server := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", config.ServerPort),
		Handler:      app.routes(), // app.routes() returns a router a.k.a server multiplexer.
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "address", "http://"+server.Addr, "environment", config.Environment)
	// Start the HTTP server.
	err = server.ListenAndServe()
	connPool.Close()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(config util.Config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct.
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, err
	}

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5-second deadline, then this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
