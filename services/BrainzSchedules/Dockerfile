FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/BrainzSchedules/go.mod services/BrainzSchedules/go.sum ./
COPY services/BrainzSchedules/go.work services/BrainzSchedules/go.work.sum ./
COPY common ./common
COPY services/BrainzSchedules ./services/BrainzSchedules

WORKDIR /app/services/BrainzSchedules

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd
# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]