FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/trip-service \
    ./cmd/service

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 tripgo && \
    adduser -D -u 1000 -G tripgo tripgo

WORKDIR /app

COPY --from=builder /bin/trip-service /app/trip-service

COPY migrations /app/migrations

RUN chown -R tripgo:tripgo /app

USER tripgo

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/trip-service"]