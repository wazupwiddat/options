# Use an official Golang runtime as a parent image
FROM golang:1.18-alpine

# Set the working directory inside the container
WORKDIR /app


# Copy the local package files to the container's workspace
COPY go.mod go.sum ./

# Copy the rest of the application's source code
COPY . .

# Build the application
RUN go build -o main .

# Run the binary program produced by `go install`
CMD ["./main"]
