FROM golang:1.25.0-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/kinetria cmd/kinetria/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/kinetria .

EXPOSE 8080

CMD ["./kinetria"]
