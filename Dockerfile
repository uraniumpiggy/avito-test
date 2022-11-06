FROM golang:alpine

WORKDIR /app/servce

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "cmd/main/main.go"]