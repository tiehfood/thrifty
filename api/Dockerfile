FROM golang:1.23.6-alpine AS builder

COPY . /app
WORKDIR /app

ENV GOOS=linux
RUN go get .; \
    go install github.com/swaggo/swag/cmd/swag@latest; \
    swag init; \
    go build -o thrifty .; \
    chmod +x thrifty;

FROM scratch
COPY --from=builder /app/thrifty /thrifty
ENTRYPOINT ["/thrifty"]
