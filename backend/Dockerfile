# Anvil Analytics backend
FROM golang:1.24

WORKDIR /app

COPY . .

RUN go build -o /usr/local/bin/anvil-server ./cmd/server

ENV ANVIL_ADDR=:8080
EXPOSE 8080

CMD ["anvil-server"]
