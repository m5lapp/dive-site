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
	addr       string
	debug      bool
	dsn        string
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
	debug              bool
	diveProperties     models.DivePropertyModelInterface
	dives              models.DiveModelInterface
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
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "HTTP network address")
	flag.BoolVar(&cfg.debug, "debug", false, "Turn on debug mode")
	flag.StringVar(&cfg.dsn, "db-dsn", "", "PostgreSQL data source name")
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

	db, err := openDB(cfg.dsn)
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
		agencies:           &models.AgencyModel{DB: db},
		agencyCourses:      &models.AgencyCourseModel{DB: db},
		buddies:            &models.BuddyModel{DB: db},
		buddyRoles:         &models.BuddyRoleModel{DB: db},
		certifications:     &models.CertificationModel{DB: db},
		countries:          &models.CountryModel{DB: db},
		currencies:         &models.CurrencyModel{DB: db},
		currents:           &models.CurrentModel{DB: db},
		diveProperties:     &models.DivePropertyModel{DB: db},
		diveSites:          &models.DiveSiteModel{DB: db},
		entryPoints:        &models.EntryPointModel{DB: db},
		equipment:          &models.EquipmentModel{DB: db},
		formDecoder:        formDecoder,
		gasMixes:           &models.GasMixModel{DB: db},
		operators:          &models.OperatorModel{DB: db},
		operatorTypes:      &models.OperatorTypeModel{DB: db},
		sessionManager:     sessionManager,
		tankConfigurations: &models.TankConfigurationModel{DB: db},
		tankMaterials:      &models.TankMaterialModel{DB: db},
		trips:              &models.TripModel{DB: db},
		users:              &models.UserModel{DB: db},
		waterBodies:        &models.WaterBodyModel{DB: db},
		waterTypes:         &models.WaterTypeModel{DB: db},
		waves:              &models.WavesModel{DB: db},
	}

	dm, err := models.NewDiveModel(db, app.equipment, app.diveProperties)
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
