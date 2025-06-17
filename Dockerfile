FROM golang:tip-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/ariand ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add curl ca-certificates

WORKDIR /app

COPY --from=builder /app/ariand /app/ariand
COPY --from=builder /app/docs ./docs

ENTRYPOINT ["/app/ariand"]
