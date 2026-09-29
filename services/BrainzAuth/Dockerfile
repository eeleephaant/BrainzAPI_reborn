FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/BrainzAuth/go.mod services/BrainzAuth/go.sum ./
COPY services/BrainzAuth/go.work ./
COPY services/BrainzAuth ./services/BrainzAuth
COPY common ./common

WORKDIR /app/services/BrainzAuth

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd
# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
