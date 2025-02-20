# Use the official Go image as the base image
FROM golang:1.21-alpine

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/main.go

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./main"]
