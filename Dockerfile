FROM danielmichaels/ci-toolkit as toolkit
FROM golang:1.22-bookworm AS builder

WORKDIR /build
# only copy mod file for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY --from=toolkit ["/usr/local/bin/goa", "/usr/local/bin/goa"]
COPY --from=toolkit ["/usr/local/bin/task", "/usr/local/bin/task"]

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

COPY . .

RUN ["task", "gen"]

RUN apt-get install git -y &&\
    go build  \
    -ldflags="-s -w" \
    -o app ./cmd/app

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder ["/build/app", "/usr/bin/app"]

RUN apt-get update && apt-get install ca-certificates curl -y

# ensures that migrations are run using the embedded files
ENV DOCKER=1
ENTRYPOINT ["app", "serve"]