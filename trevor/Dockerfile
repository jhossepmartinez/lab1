FROM golang:1.23.0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /main

EXPOSE 50053
ENV RABBITMQ_HOST=10.35.168.23
CMD ["/main"]

