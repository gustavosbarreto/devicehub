# base stage
FROM golang:1.22.6-alpine3.19 AS base

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./
COPY ./api/go.mod ./api/go.mod

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/adm

COPY ./adm/go.mod ./adm/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

COPY ./api $GOPATH/src/github.com/shellhub-io/shellhub/api
COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./adm .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/adm

RUN go build

# development stage
FROM builder AS development

RUN apk add --update openssl build-base docker-cli
RUN go install github.com/cosmtrek/air@v1.51.0 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./adm/entrypoint-dev.sh /entrypoint-dev.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/adm

RUN go build

ENTRYPOINT ["/entrypoint-dev.sh"]

# production stage
FROM alpine:3.20.3 AS production

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/adm/adm /adm
