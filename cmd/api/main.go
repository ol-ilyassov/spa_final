package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"github.com/ol-ilyassov/spa_final/internal/data"
	"github.com/ol-ilyassov/spa_final/internal/jsonlog"
	"github.com/ol-ilyassov/spa_final/internal/mailer"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Application Version number
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
	limiter struct {
		rps     float64 // Request per second
		burst   int     // Number of maximum request in single burst
		enabled bool    // Is RateLimiter turned On
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// Dependencies for HTTP handlers, helpers, and middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// Instance of config struct
	var cfg config

	// Default values
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Env variable
	//flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	//flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:123@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "8db0d4c74dccb8", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "8376f58c61e62a", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "RIG <no-reply@rig.mail.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create Connection Pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// Npw: cmdline, memstats, version.
	expvar.NewString("version").Set(version)

	// Num of Active goroutines
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	// DB pool Statistics
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	// Current Unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Instance of application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
