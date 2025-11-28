FROM docker.io/golang:1.25.3-trixie AS builder

WORKDIR /go/src/app/

# The Golang toolchain image will take the ARGs passed from BuildKit and assign
# them automatically to the GOARCH and GOOS environment variables.
ARG TARGETARCH
ARG TARGETOS

ENV CGO_ENABLED="0"
ENV GOFLAGS="--buildvcs=false"
ENV MIGRATE_VERSION="4.19.0"
ENV MIGRATE_URL="https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.${GOOS}-${GOARCH}.tar.gz"

RUN echo $MIGRATE_URL && curl -L "${MIGRATE_URL}" | tar -xvz -C /go/bin/ migrate

COPY . /go/src/app/

RUN make build/web

################################################################################

FROM gcr.io/distroless/static-debian13

COPY ./migrations/ /migrations/

COPY --from=builder /go/bin/migrate /migrate

COPY --from=builder /go/src/app/bin/web /dive-site

EXPOSE 8080

ENTRYPOINT ["/dive-site"]

