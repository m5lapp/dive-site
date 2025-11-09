FROM docker.io/golang:1.25.3-trixie AS builder

WORKDIR /go/src/app/

ENV CGO_ENABLED="0"
ENV GOFLAGS="--buildvcs=false"

COPY . /go/src/app/

RUN make build/web

################################################################################

FROM gcr.io/distroless/static-debian13

COPY --from=builder /go/src/app/bin/web /dive-site

EXPOSE 8080

ENTRYPOINT ["/dive-site"]

