FROM golang:1.23.0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /main

EXPOSE 50051
ENV RABBITMQ_HOST=amqp://admin:admin@192.168.1.6:5673/
CMD ["/main"]

