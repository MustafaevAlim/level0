FROM golang:1.24-alpine AS builder
WORKDIR /app



COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache curl tar \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz \
    | tar xvz -C /usr/local/bin \
    && chmod +x /usr/local/bin/migrate



COPY . .
COPY .env .

RUN go build -mod=vendor -o app ./cmd/myapp/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/.env .
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

COPY scripts/entrypoint.sh /root/entrypoint.sh
RUN chmod +x /root/entrypoint.sh

ENTRYPOINT ["/root/entrypoint.sh"]

