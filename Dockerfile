FROM golang:1.21.5-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o /bin/auth-proxy ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/bin/auth-proxy .

CMD [ "./auth-proxy" ]


