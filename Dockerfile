# Start from a Debian-based image with Go installed
FROM golang:1.17

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod ./
COPY go.sum ./

# Download necessary Go modules
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main .

# Run the application
CMD ["./main"]