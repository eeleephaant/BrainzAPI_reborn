FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/BrainzDevelopers/go.work services/BrainzDevelopers/go.work.sum ./

COPY common ./common
COPY services/BrainzDevelopers ./services/BrainzDevelopers

WORKDIR /app/services/BrainzDevelopers

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-w -s" -o /app/main ./cmd

FROM alpine:latest

WORKDIR /root
COPY --from=builder /app/main .
CMD ["./main"]
