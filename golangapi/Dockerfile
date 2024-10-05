# Use Golang image as the base
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files and download the dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o main .

# Expose the port the app runs on
EXPOSE 8080

# Start the app
CMD ["./main"]
