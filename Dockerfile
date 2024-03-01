FROM golang:1.22.0-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o ./bin/auth-proxy ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/bin/auth-proxy .

CMD [ "./auth-proxy" ]


