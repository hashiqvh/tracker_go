# Set the base image
FROM golang:1.20.0-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module dependency list and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port used by the application
EXPOSE 3000

# Start the application
CMD ["./main"]
