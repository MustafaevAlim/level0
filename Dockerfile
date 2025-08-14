FROM golang:1.24-alpine AS builder
WORKDIR /app



COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct



COPY . .
COPY .env .

RUN go build -mod=vendor -o app ./cmd/myapp/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/.env .
COPY --from=builder /app/web ./web
RUN ls -a
CMD ["./app"]
