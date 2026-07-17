# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy Go module files first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o validator .

# Stage 2: Distroless runtime image for minimal attack surface
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /app/validator .

# Webhooks require TLS, we expose 8443
EXPOSE 8443

USER nonroot:nonroot
ENTRYPOINT ["./validator"]
