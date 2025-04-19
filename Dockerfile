# syntax=docker/dockerfile:1

FROM golang:1.21

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go app
RUN go build -o server ./cmd/server

# Run the server binary
CMD ["./server"]
