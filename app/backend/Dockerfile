# Usage Analytics backend
FROM golang:1.24

WORKDIR /app

COPY . .

RUN go build -o /usr/local/bin/metrics-server ./cmd/server

ENV METRICS_ADDR=:8080
EXPOSE 8080

CMD ["metrics-server"]
