package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Declare a string containing the application version number.
const version = "1.0.0"

// Declare a config struct to hold all the configuration settings for our application.
type config struct {
	port int
	env  string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middlewares.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var conf config

	// Read the value of the port and env command-line flags into the config struct.
	// Default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&conf.port, "port", 4000, "API server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: conf,
		logger: logger,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", conf.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on http://%s", conf.env, server.Addr)
	err := server.ListenAndServe()
	logger.Fatal(err)
}
