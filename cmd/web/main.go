package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/m5lapp/divesite-monolith/internal/models"

	_ "github.com/lib/pq"
)

type config struct {
	addr  string
	debug bool
	db    struct {
		dsn             string
		maxOpenConns    int
		maxIdleConns    int
		maxConnLifetime time.Duration
		maxConnIdleTime time.Duration
		timeouts        models.QueryTimeouts
	}
	termPeriod time.Duration
	tlsCert    string
	tlsKey     string
}

func (c config) validate(logger *slog.Logger) {
	if c.termPeriod < 1*time.Second || c.termPeriod > 300*time.Second {
		logger.Error(
			"The termination shutdown grace period must be between 1 and 300 seconds",
			"--term-period",
			c.termPeriod.String(),
		)
		os.Exit(1)
	}

	if (c.tlsCert == "" && c.tlsKey != "") || (c.tlsCert != "" && c.tlsKey == "") {
		logger.Error(
			"The --tls-cert and --tls-key flags are mutually inclusive and must both be provided to use TLS",
			"--tls-cert",
			c.tlsCert,
			"--tls-key",
			c.tlsKey,
		)
		os.Exit(1)
	}
}

type app struct {
	agencies           models.AgencyModelInterface
	agencyCourses      models.AgencyCourseModelInterface
	buddies            models.BuddyModelInterface
	buddyRoles         models.BuddyRoleModelInterface
	certifications     models.CertificationModelInterface
	config             config
	countries          models.CountryModelInterface
	currencies         models.CurrencyModelInterface
	currents           models.CurrentModelInterface
	diveProperties     models.DivePropertyModelInterface
	dives              models.DiveModelInterface
	divePlans          models.DivePlanModelInterface
	diveSites          models.DiveSiteModelInterface
	entryPoints        models.EntryPointModelInterface
	equipment          models.EquipmentModelInterface
	formDecoder        *form.Decoder
	gasMixes           models.GasMixModelInterface
	log                *slog.Logger
	operators          models.OperatorModelInterface
	operatorTypes      models.OperatorTypeModelInterface
	tankConfigurations models.TankConfigurationModelInterface
	tankMaterials      models.TankMaterialModelInterface
	templateCache      map[string]*template.Template
	trips              models.TripModelInterface
	users              models.UserModelInterface
	sessionManager     *scs.SessionManager
	waterBodies        models.WaterBodyModelInterface
	waterTypes         models.WaterTypeModelInterface
	waves              models.WavesModelInterface
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxLifetime(cfg.db.maxConnLifetime)
	db.SetConnMaxIdleTime(cfg.db.maxConnIdleTime)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "HTTP network address")
	flag.BoolVar(&cfg.debug, "debug", false, "Turn on debug mode")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL data source name")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 0, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 0, "PostgreSQL max idle connections")
	flag.DurationVar(
		&cfg.db.maxConnLifetime,
		"db-max-conn-lifetime",
		0*time.Second,
		"PostgreSQL max connection idle time",
	)
	flag.DurationVar(
		&cfg.db.maxConnIdleTime,
		"db-max-conn-idle-time",
		0*time.Second,
		"PostgreSQL max connection idle time",
	)
	flag.DurationVar(
		&cfg.db.timeouts.Quick,
		"query-timeout-quick",
		1*time.Second,
		"DB timeout for quick, simple queries",
	)
	flag.DurationVar(
		&cfg.db.timeouts.Standard,
		"query-timeout-standard",
		3*time.Second,
		"DB timeout for standard queries",
	)
	flag.DurationVar(
		&cfg.db.timeouts.Moderate,
		"query-timeout-moderate",
		5*time.Second,
		"DB timeout for more demanding queries",
	)
	flag.DurationVar(
		&cfg.db.timeouts.Complex,
		"query-timeout-complex",
		10*time.Second,
		"DB timeout for complex queries",
	)
	flag.DurationVar(
		&cfg.db.timeouts.Bulk,
		"query-timeout-bulk",
		20*time.Second,
		"DB timeout for large, bulk queries",
	)
	flag.DurationVar(&cfg.termPeriod, "term-period", 30*time.Second, "Termination grace period")
	flag.StringVar(&cfg.tlsCert, "tls-cert", "", "TLS cert file path if TLS is required")
	flag.StringVar(&cfg.tlsKey, "tls-key", "", "TLS key file path if TLS is required")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg.validate(logger)

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(3)
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
		config:             cfg,
		log:                logger,
		templateCache:      templateCache,
		agencies:           &models.AgencyModel{DB: db, Timeouts: cfg.db.timeouts},
		agencyCourses:      &models.AgencyCourseModel{DB: db, Timeouts: cfg.db.timeouts},
		buddies:            &models.BuddyModel{DB: db, Timeouts: cfg.db.timeouts},
		buddyRoles:         &models.BuddyRoleModel{DB: db, Timeouts: cfg.db.timeouts},
		certifications:     &models.CertificationModel{DB: db, Timeouts: cfg.db.timeouts},
		countries:          &models.CountryModel{DB: db, Timeouts: cfg.db.timeouts},
		currencies:         &models.CurrencyModel{DB: db, Timeouts: cfg.db.timeouts},
		currents:           &models.CurrentModel{DB: db, Timeouts: cfg.db.timeouts},
		diveProperties:     &models.DivePropertyModel{DB: db, Timeouts: cfg.db.timeouts},
		divePlans:          &models.DivePlanModel{DB: db, Timeouts: cfg.db.timeouts},
		diveSites:          &models.DiveSiteModel{DB: db, Timeouts: cfg.db.timeouts},
		entryPoints:        &models.EntryPointModel{DB: db, Timeouts: cfg.db.timeouts},
		equipment:          &models.EquipmentModel{DB: db, Timeouts: cfg.db.timeouts},
		formDecoder:        formDecoder,
		gasMixes:           &models.GasMixModel{DB: db, Timeouts: cfg.db.timeouts},
		operators:          &models.OperatorModel{DB: db, Timeouts: cfg.db.timeouts},
		operatorTypes:      &models.OperatorTypeModel{DB: db, Timeouts: cfg.db.timeouts},
		sessionManager:     sessionManager,
		tankConfigurations: &models.TankConfigurationModel{DB: db, Timeouts: cfg.db.timeouts},
		tankMaterials:      &models.TankMaterialModel{DB: db, Timeouts: cfg.db.timeouts},
		trips:              &models.TripModel{DB: db, Timeouts: cfg.db.timeouts},
		users:              &models.UserModel{DB: db, Timeouts: cfg.db.timeouts},
		waterBodies:        &models.WaterBodyModel{DB: db, Timeouts: cfg.db.timeouts},
		waterTypes:         &models.WaterTypeModel{DB: db, Timeouts: cfg.db.timeouts},
		waves:              &models.WavesModel{DB: db, Timeouts: cfg.db.timeouts},
	}

	dm, err := models.NewDiveModel(db, cfg.db.timeouts, app.equipment, app.diveProperties)
	if err != nil {
		app.log.Error("Could not instantiate DiveModel: " + err.Error())
		os.Exit(4)
	}
	app.dives = dm

	err = app.serve()

	if err != nil {
		app.log.Error(err.Error())
		os.Exit(5)
	}
}
