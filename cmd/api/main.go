package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Application Version Number
const version = "1.0.0"

// Configuration Settings
type config struct {
	port int    // Network Port
	env  string // Current Operating Environment
	db   struct {
		dsn          string // Database Connection
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Dependencies for HTTP handlers, helpers, and middleware
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// Instance of config struct
	var cfg config

	// Default values
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Env variable
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	//flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Create Connection Pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Printf("database connection pool established")

	// Instance of application struct
	app := &application{
		config: cfg,
		logger: logger,
	}

	// HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// Returns a connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Create an empty connection pool using DSN.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Value less than or equal 0 means no limit.
	// Set the maximum number of open (in-use + idle) connections in the pool.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	// Set the maximum number of idle connections in the pool.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Establish a new connection to the database. If no connection in 5 sec period,
	// then retrieve error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
