FROM golang:1.23.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /lb

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /lb .

EXPOSE 8080

CMD ["./lb"]