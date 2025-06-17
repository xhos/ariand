FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ariand ./cmd/

FROM gcr.io/distroless/static-debian11:latest

WORKDIR /app
COPY --from=curlimages/curl:latest /usr/bin/curl /usr/bin/curl
COPY --from=builder /app/ariand /app/ariand
ENTRYPOINT ["/app/ariand"]