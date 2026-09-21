FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /url-shortener ./cmd/api

FROM alpine:3.21

COPY --from=build /url-shortener /usr/local/bin/url-shortener

EXPOSE 8080
ENTRYPOINT ["url-shortener"]
