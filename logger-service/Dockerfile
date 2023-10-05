#FROM golang:1.21.0-alpine3.18 AS builder
#WORKDIR /app
#COPY . .
#RUN go build -o brokerapp ./cmd/api

#RUN chmod +x /app/brokerapp

FROM alpine:3.18
WORKDIR /app
COPY loggerApp /app

CMD [ "/app/loggerApp" ]