# Dive Site - A Web Portal For SCUBA Diving Written in Go

> This is a work in progress and will be updated over time.

Dive Site is a web portal for SCUBA diving which is primarily written in Go and uses a PostgreSQL database. Currently it allows users to log dives including information about their buddies, certifications/courses, trips like liveaboards and dive operators like dive schools and clubs. More features are planned in the future, including dive planning.

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

