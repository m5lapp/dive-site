package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/m5lapp/divesite-monolith/internal/models"

	_ "github.com/lib/pq"
)

type app struct {
	log            *slog.Logger
	templateCache  map[string]*template.Template
	agencies       models.AgencyModelInterface
	buddies        models.BuddyModelInterface
	buddyRoles     models.BuddyRoleModelInterface
	countries      models.CountryModelInterface
	diveSites      models.DiveSiteModelInterface
	operators      models.OperatorModelInterface
	operatorTypes  models.OperatorTypeModelInterface
	users          models.UserModelInterface
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	waterBodies    models.WaterBodyModelInterface
	waterTypes     models.WaterTypeModelInterface
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")
	dsn := flag.String("db-dsn", "", "PostgreSQL data source name")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}
	defer db.Close()

	formDecoder := form.NewDecoder()
	FormDecoderRegisterTimeType(formDecoder, nil)
	FormDecoderRegisterTimeLocationType(formDecoder)

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// Setting this means that the cookie will only be sent by the user's web
	// browser if there is a TLS connection.
	sessionManager.Cookie.Secure = true

	app := app{
		log:            logger,
		templateCache:  templateCache,
		agencies:       &models.AgencyModel{DB: db},
		buddies:        &models.BuddyModel{DB: db},
		buddyRoles:     &models.BuddyRoleModel{DB: db},
		countries:      &models.CountryModel{DB: db},
		diveSites:      &models.DiveSiteModel{DB: db},
		users:          &models.UserModel{DB: db},
		formDecoder:    formDecoder,
		operators:      &models.OperatorModel{DB: db},
		operatorTypes:  &models.OperatorTypeModel{DB: db},
		sessionManager: sessionManager,
		waterBodies:    &models.WaterBodyModel{DB: db},
		waterTypes:     &models.WaterTypeModel{DB: db},
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:      app.routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  5 * time.Second,
		TLSConfig:    tlsConfig,
		WriteTimeout: 10 * time.Second,
	}

	app.log.Info("Starting server", "addr", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	app.log.Error(err.Error())
	os.Exit(3)
}
