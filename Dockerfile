# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build API server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /bin/samaj-api ./cmd/api

# Build Worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /bin/samaj-worker ./cmd/worker

# ============================================
# Stage 2: API Runtime (<25MB)
# ============================================
FROM gcr.io/distroless/static-debian12 AS api

COPY --from=builder /bin/samaj-api /samaj-api

EXPOSE 8080

ENTRYPOINT ["/samaj-api"]

# ============================================
# Stage 3: Worker Runtime (<25MB)
# ============================================
FROM gcr.io/distroless/static-debian12 AS worker

COPY --from=builder /bin/samaj-worker /samaj-worker

EXPOSE 8081

ENTRYPOINT ["/samaj-worker"]
