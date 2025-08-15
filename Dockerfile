# Build
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /people-api ./cmd/api

# Run
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /people-api /people-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/people-api"]
