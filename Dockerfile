FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download deps first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o nomvox-server ./cmd/server

# ── Minimal runtime image ───────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/nomvox-server .

EXPOSE 8080

CMD ["./nomvox-server"]
