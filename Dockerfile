FROM golang:alpine as builder
WORKDIR /app/
ADD . .
RUN go build -o ./gonpm ./cmd

FROM alpine:latest
RUN apk update && apk --no-cache add ca-certificates
COPY --from=builder /app/gonpm /app/gonpm

EXPOSE 8080
ENTRYPOINT [ "/app/gonpm" ]
