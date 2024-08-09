package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/mailer"
)

// Application version number
const version = "1.0.0"

// Configuration settings for our application.
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	smtp struct {
		username    string
		password    string
		senderName  string
		senderEmail string
	}
}

// application hold dependencies for our HTTP handlers, helpers, and middlewares.
type application struct {
	config config
	logger *slog.Logger
	store  db.Store
	mailer mailer.EmailSender
}

func main() {
	var cfg config

	// Read the value of the port and env command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	flag.StringVar(&cfg.smtp.username, "mailtrap-smtp-username", os.Getenv("MAILTRAP_SMTP_USERNAME"), "Mailtrap SMTP username")
	flag.StringVar(&cfg.smtp.password, "mailtrap-smtp-password", os.Getenv("MAILTRAP_SMTP_PASSWORD"), "Mailtrap SMTP password")
	flag.StringVar(&cfg.smtp.senderName, "smtp-sender-name", "Greenlight", "SMTP sender name")
	flag.StringVar(&cfg.smtp.senderEmail, "smtp-sender-email", "no-reply@greenlight.katatrina.net", "SMTP sender email address")

	flag.Parse()

	// Initialize a new structured logger which writes log entries to the standard out stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	connPool, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer connPool.Close()

	logger.Info("database connection pool established")

	store := db.NewStore(connPool)

	app := application{
		config: cfg,
		logger: logger,
		store:  store,
		mailer: mailer.NewMailtrapSender(cfg.smtp.username, cfg.smtp.password, cfg.smtp.senderName, cfg.smtp.senderEmail),
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
	logger.Info("server is listening", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB creates a new connection pool to our PostgreSQL database.
func openDB(cfg config) (*pgxpool.Pool, error) {
	connPool, err := pgxpool.New(context.Background(), cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	// Create a context with a 5-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = connPool.Ping(ctx)
	if err != nil {
		defer connPool.Close()
		return nil, err
	}

	return connPool, nil
}
