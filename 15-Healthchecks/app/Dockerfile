# Use the official Golang image
FROM golang:alpine AS builder
WORKDIR /app

# Copy and build the application
COPY . .
RUN go build -o myapp main.go

# Use a minimal base image for the final container
FROM scratch
WORKDIR /root/
COPY --from=builder /app/myapp .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./myapp"]
