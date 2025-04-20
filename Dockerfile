# syntax=docker/dockerfile:1

FROM golang:1.24.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

RUN templ generate

RUN go build -o server ./cmd/server

CMD ["./server"]
