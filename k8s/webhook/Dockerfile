# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Initialize go module
RUN go mod init webhook
RUN go get k8s.io/api/admission/v1
RUN go get k8s.io/api/core/v1
RUN go get k8s.io/apimachinery/pkg/apis/meta/v1

COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o validator .

# Stage 2: Create lightweight runtime image
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/validator .

# Webhooks require TLS, we expose 8443
EXPOSE 8443

ENTRYPOINT ["./validator"]
