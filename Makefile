# The dash here ignores any errors if the included file does not exist.
-include .env

# ============================================================================ #
# HELPERS
# ============================================================================ #

## help: Print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# ============================================================================ #
# DEVELOPMENT
# ============================================================================ #

## update-bootstrap version=X.Y.Z: Update the version of Bootstrap
.PHONY: update-bootstrap
update-bootstrap:
	@echo "Downloading Bootstrap files at version ${version}..."
	curl -Lo ui/static/js/bootstrap.bundle.min.js \
		https://cdn.jsdelivr.net/npm/bootstrap@${version}/dist/js/bootstrap.bundle.min.js
	curl -Lo ui/static/css/bootstrap.min.css \
		https://cdn.jsdelivr.net/npm/bootstrap@${version}/dist/css/bootstrap.min.css
	@echo ""
	@echo "Remember to also update the integrity hash of each file in ui/html/base.tmpl,"
	@echo "the hash values can be found at https://getbootstrap.com/docs/5.3/getting-started/introduction/"

## run: Run the cmd/web application
.PHONY: run
run:
	go run ./cmd/web/ \
		--addr ":8080" \
		--db-dsn ${DIVESITE_DB_DSN} \
		--debug=true

## run/tls: Run the cmd/web application using TLS
.PHONY: run/tls
run/tls:
	go run ./cmd/web/ \
		--addr ":8080" \
		--db-dsn ${DIVESITE_DB_DSN} \
		--debug=true \
		--tls-cert ./tls/cert.pem \
		--tls-key ./tls/key.pem

## gen-cert: Generate a TLS certificate and key for testing on localhost
.PHONY: gen-cert
gen-cert:
	@echo "Generating new TLS certificate and key for testing on localhost..."
	go run $$(dirname $$(dirname $$(which go)))/src/crypto/tls/generate_cert.go \
		--host localhost \
		--rsa-bits 2048
	mkdir tls/ || true
	mv cert.pem key.pem tls/
	@echo "New TLS certificate and key written to tls/ directory"

# ============================================================================ #
# DATABASE
# ============================================================================ #

## db/connect: connect to the database using psql
.PHONY: db/connect
db/connect:
	psql ${DIVESITE_DB_DSN}

## db/run: Run SQL command=$1 against the database.
.PHONY: db/run command=$1
db/run:
	psql ${DIVESITE_DB_DSN} \
		-c "${command}"

## db/start/integration: start a local database for running integration tests
.PHONY: db/start/integration
db/start/integration:
	@echo "Starting new database instance for integration testing..."
	podman container run \
		-d --rm \
		--name divesite-integration-test-db \
		-e POSTGRES_USER=divesite_integration_test \
		-e POSTGRES_PASSWORD=password \
		-p 5432:5432 \
		docker.io/postgres:14-alpine

## db/stop/integration: stop the local database for running integration tests
.PHONY: db/stop/integration
db/stop/integration:
	@echo "Stop the database instance for integration testing..."
	podman container stop divesite-integration-test-db

## db/migrations/new name=$1: Create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating migration files for ${name}..."
	migrate create --seq --ext .sql --dir ./migrations/ ${name}

## db/migrations/up n=$1: Apply all or N up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running ${n} up migrations..."
	@# For some reason, migrate requires sslmode=disable in the DSN string.
	migrate \
		--path ./migrations/ \
		--database ${DIVESITE_DB_DSN}?sslmode=disable \
		up ${n}

## db/migrations/down n=$1: Apply all or N down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo "Running ${n} down migrations..."
	@# For some reason, migrate requires sslmode=disable in the DSN string.
	migrate \
		--path ./migrations/ \
		--database ${DIVESITE_DB_DSN}?sslmode=disable \
		down ${n}

## db/migrations/clean: Set the migration status to clean for retrying.
.PHONY: db/migrations/clean
db/migrations/clean:
	@echo "Setting the migration status to clean..."
	@# For some reason, migrate requires sslmode=disable in the DSN string.
	psql ${DIVESITE_DB_DSN} \
		-c 'update schema_migrations set dirty = false returning *;'

# ============================================================================ #
# QUALITY CONTROL
# ============================================================================ #

## audit: Tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "Running tests..."
	go test --race --vet off ./...

## vendor: Tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Vendoring dependencies..."
	go mod vendor

# ============================================================================ #
# BUILD
# ============================================================================ #

## build/web: Build the cmd/web application
.PHONY: build/web
build/web:
	@echo "Building cmd/web"
	go build --ldflags "-s" -o ./bin/web ./cmd/web

