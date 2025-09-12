FROM golang:1.23.0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /main

EXPOSE 50052
ENV LESTER_HOST=10.35.168.23
ENV TREVOR_HOST=10.35.168.25
ENV FRANKLIN_HOST=10.35.168.26
CMD ["/main"]
