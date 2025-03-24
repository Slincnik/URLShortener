FROM golang:1.24-alpine3.20 as build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /urlshortener ./cmd/shortener

FROM alpine:3.20 as build-release

WORKDIR /

COPY --from=build-stage /urlshortener /urlshortener

EXPOSE 8080

ENTRYPOINT ["/urlshortener"]

HEALTHCHECK CMD curl --fail http://localhost/health || exit 1