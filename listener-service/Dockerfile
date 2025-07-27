# Step 1: Build the Go binary
FROM golang:1.23.3-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY cmd ./cmd
COPY event ./event
COPY internal ./internal

# Build the Go app
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./listenerApp

# Step 2: Create a minimal runtime image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the builder image
COPY --from=builder /app/cmd/listenerApp .

# Command to run the executable
CMD ["./listenerApp"]
