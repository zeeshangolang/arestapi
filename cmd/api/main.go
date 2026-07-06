package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"main/internal/data"
	"main/internal/mailer"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn         string
		maxOpenConn int
		maxIdleConn int
		MaxIdleTime time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabeld bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "PORT NETWORM")
	flag.StringVar(&cfg.env, "env", "development", "ENVIzroment(deve|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN_NOSSL"), "POSTQRESl DSN")
	flag.IntVar(&cfg.db.maxOpenConn, "db-maxopenconn", 25, "maximum open connections")
	flag.IntVar(&cfg.db.maxIdleConn, "db-maxIdlecon", 25, "maximum idle connections")
	flag.DurationVar(&cfg.db.MaxIdleTime, "db-maxIdleTime", 15*time.Minute, "Maximum Idle timeout ")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "request-per-second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "maximum-request-burst")
	flag.BoolVar(&cfg.limiter.enabeld, "limite-enalbled", true, "enable the limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "live.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "api", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@kineta.site>", "SMTP sender")

	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := OpenDb(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	err = app.server()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

}

func OpenDb(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConn)
	db.SetMaxIdleConns(cfg.db.maxIdleConn)
	db.SetConnMaxIdleTime(cfg.db.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
