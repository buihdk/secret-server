FROM golang:1.12-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o secretserver .

FROM alpine:3.9
WORKDIR /app
COPY --from=builder /app/secretserver .
EXPOSE 8080
CMD ["./secretserver"]
