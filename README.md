# Dive Site - A Web Portal For SCUBA Diving Written in Go

> This is a work in progress and will be updated over time.

Dive Site is a web portal for SCUBA diving which is primarily written in Go and uses a PostgreSQL database. Currently it allows users to log dives including information about their buddies, certifications/courses, trips like liveaboards and dive operators like dive schools and clubs. It also allows for dive planning using the [m5lapp/diveplanner](https://github.com/m5lapp/diveplanner) library and will chart the profile of any planned dives and calculate the DSR table.

It was mainly created as a learning exercise and if you are looking for an open-source tool to log your SCUBA dives, I would suggest the excellent [Subsurface](https://subsurface-divelog.org/) created by Linus Torvalds and many others.

## Getting Started

### Local Development

```bash
# Create a local environment file to store your database credentials.
cat > .env << EOF
    # %25 in the password here represents the percent sign itself, %40 represents the
    # @ character.
    export DIVESITE_DB_DSN='postgres://USERNAME:PASSWORD@HOST/SERVICE'
EOF

source .env

# Run the database migrations to set up the required structure.
make db/migrations/up

# Run the application locally.
make run

# Alternatively, you can view the available command-line flags as follows.
go run ./cmd/web/ --help
```

After this, you should be able to navigate to [https://localhost:8080/user/sign-up] and register an account that you can use for testing.

### Deployment

A container image is also available for deploying in containerised environments as follows using your chosen container runtime (e.g. `podman`):

```bash
export DIVESITE_DB_DSN="postgresql://divesite_user:Sup3rS3cr3tP455w0rd@db.example.com:5432/divesite_prod"

# Override the entrypoint to run the database migrations.
podman container run \
    --rm \
    --entrypoint /migrate \
    ghcr.io/m5lapp/dive-site:vX.Y.Z \
    --path /migrations/ \
    --database ${DIVESITE_DB_DSN} \
    up

# Once the migrations have run, 
podman container run \
    --rm \
    ghcr.io/m5lapp/dive-site:vX.Y.Z \
    --addr ":8080" \
    --db-dsn ${DIVESITE_DB_DSN}
```

Alternatively, for deployment on Kubernetes, there is a Helm Chart available from [m5lapp/helm-charts](https://github.com/m5lapp/helm-charts/tree/main/charts/dive-site) which can be used as follows. This assumes you have created a `values.yaml` file to set the values you want to override from the chart's [default values](https://github.com/m5lapp/helm-charts/blob/main/charts/dive-site/values.yaml) file.

```bash
# Add the Helm Chart repository.
helm repo add m5lapp https://m5lapp.github.io/helm-charts

# Create a namespace.
kubectl create namespace dive-site

# Install the Helm Chart into the new namespace.
helm install dive-site m5lapp/dive-site \
    --namespace dive-site \
    --values values.yaml
```

Every time a new Pod starts up it will first launch an init container and attempt to run any new database migrations automatically.

