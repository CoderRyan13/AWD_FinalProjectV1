//Filename: cmd/api/main.go

package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"sync"
	"time"

	"AWD_FinalProject.ryanarmstrong.net/internal/data"
	"AWD_FinalProject.ryanarmstrong.net/internal/jsonlog"
	"AWD_FinalProject.ryanarmstrong.net/internal/mailer"
	_ "github.com/lib/pq"
)

// The application version number
const version = "1.0.0"

// The configuration settings
type config struct {
	port int
	env  string // development, staging, production, etc.
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64 // request/second
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Dependency Injection
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	// read in the flags that are needed to populate our config
	flag.IntVar(&cfg.port, "port", 4001, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FORUM_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle connections time")
	// These are flags for the rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	// These are flags for the mailer
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "6b26d672d3a39f", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "21df5a4c1c7ee2", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "G-Forum <no-reply@Gforum.ryanarmstrong.net>", "SMTP sender")

	flag.Parse()
	// Create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	// Create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	// Log the successful connection pool
	logger.PrintInfo("database connection pool established", nil)
	// Create an instance of our application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// The openDB() function returns a *sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// Create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
