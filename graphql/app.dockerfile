# Build Stage
FROM golang:1.23.5-alpine AS build

# Install necessary dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set up working directory
WORKDIR /go/src/github.com/Koch13o1/go-grpc-graphql-microservice

# Copy module files and download dependencies first (leveraging Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY . .

# Build the catalog microservice binary
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./graphql

# Final Runtime Image
FROM alpine:3.11

# Set up working directory
WORKDIR /usr/bin

# Copy built application from the build stage
COPY --from=build /go/bin/app .

# Expose the service port
EXPOSE 8080

# Run the catalog service
CMD ["./app"]
