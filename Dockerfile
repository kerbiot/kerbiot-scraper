FROM golang:1.19 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /app/kerbiot-scraper .

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/kerbiot-scraper /app/kerbiot-scraper

ENTRYPOINT [ "/app/kerbiot-scraper" ]