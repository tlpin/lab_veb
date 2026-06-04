FROM golang:1.26-alpine

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init

RUN go build -o main .

EXPOSE 4200

CMD ["./main"]