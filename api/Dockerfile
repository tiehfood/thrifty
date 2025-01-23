FROM golang:1.23.4-alpine AS builder

COPY . /app
WORKDIR /app

ENV GOOS=linux
RUN go get .; \
    go build -o thrifty .; \
    chmod +x thrifty;

FROM scratch
COPY --from=builder /app/thrifty /thrifty
ENTRYPOINT ["/thrifty"]
