package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/katatrina/greenlight/util"
	"log"
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
	logger *log.Logger
	config util.Config
}

func main() {
	config, err := util.LoadConfig("./app_config.yaml")

	// Read the value of the port and env command-line flags into the config struct.
	// Default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&config.ServerPort, "port", 4000, "API server port")
	flag.StringVar(&config.Environment, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(config)
	if err != nil {
		logger.Fatal(err)
	}

	db.Close()

	app := &application{
		logger: logger,
		config: config,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", config.ServerPort),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on http://%s", config.Environment, server.Addr)
	err = server.ListenAndServe()
	logger.Fatal(err)
}

func openDB(config util.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DBDsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
