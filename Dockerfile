FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/roognis ./cmd/server

# ── Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

COPY --from=builder /bin/roognis /bin/roognis

USER appuser
EXPOSE 8080

ENTRYPOINT ["/bin/roognis"]
