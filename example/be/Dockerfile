## Build
FROM golang:alpine AS build

ENV CGO_ENABLED=1

RUN apk add --no-cache \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev

WORKDIR /

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -ldflags='-s -w -extldflags "-static"' -o /app ./

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /app /app

EXPOSE 9999/tcp

USER nonroot:nonroot

ENTRYPOINT ["/app"]