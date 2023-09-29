FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o second-microservice

FROM golang:1.20
COPY --from=builder /app/second-microservice /second-microservice
COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/second-microservice"]