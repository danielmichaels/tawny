FROM danielmichaels/ci-toolkit as toolkit
FROM node:lts-slim as node

COPY --from=toolkit ["/usr/local/bin/task", "/usr/local/bin/task"]

# PNPM is required to build the assets
RUN corepack enable pnpm
RUN mkdir -p /build
WORKDIR /build

COPY . .
RUN ["task", "assets"]

FROM golang:1.22-bookworm AS builder

WORKDIR /build
# only copy mod file for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY --from=node ["/build/assets/static/css/theme.css", "/build/assets/static/css/theme.css"]
COPY --from=node ["/build/assets/static/js/bundle.js", "/build/assets/static/js/bundle.js"]
COPY --from=toolkit ["/usr/local/bin/goa", "/usr/local/bin/goa"]
COPY --from=toolkit ["/usr/local/bin/templ", "/usr/local/bin/templ"]
COPY --from=toolkit ["/usr/local/bin/task", "/usr/local/bin/task"]

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

COPY . .

RUN ["task", "templgen"]
RUN ["task", "gen"]

RUN apt-get install git -y &&\
    go build  \
    -ldflags="-s -w" \
    -o app ./cmd/app

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=toolkit ["/usr/local/bin/goose", "/usr/local/bin/goose"]
COPY --from=builder ["/build/entrypoint", "/app/entrypoint"]
COPY --from=builder ["/build/assets/migrations", "/app/migrations"]
COPY --from=builder ["/build/app", "/usr/bin/app"]

RUN apt-get update && apt-get install ca-certificates curl -y

# ensures that migrations are run using the embedded files
ENV DOCKER=1
ENTRYPOINT ["app", "serve"]