# Step 1: Build the Go binary
FROM golang:1.24.4-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY api ./api
COPY cmd ./cmd
COPY data ./data
COPY logs ./logs
COPY internal ./internal

# Build the Go app
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./loggerApp

# Step 2: Create a minimal runtime image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the builder image
COPY --from=builder /app/cmd/loggerApp .

# Command to run the executable
CMD ["./loggerApp"]
