FROM golang:1.23.0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /main

EXPOSE 50052
CMD ["/main"]

