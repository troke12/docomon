FROM golang:1.20-alpine as builder

LABEL version="1.0"
LABEL description="Docker Container Monitoring - A simple way to monitor your docker containers, alerts via discord and google chat."
LABEL maintainer="jame_glove@yahoo.com"

RUN apk add --update --no-cache build-base
RUN apk add --update --no-cache upx

WORKDIR /app

COPY . .

RUN GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -o ./main /app/cmd/main.go

RUN upx ./main

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /app/main

CMD ["/app/main"]
