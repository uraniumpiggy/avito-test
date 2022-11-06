FROM golang:alpine

WORKDIR /app/servce

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

CMD ["go", "run", "cmd/main/main.go"]