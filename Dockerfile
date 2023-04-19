# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and build dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port on which the application will run
EXPOSE 8080

# Run the application
CMD ["./main"]
