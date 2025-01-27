# Stage 1: Build
FROM golang:alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app
COPY . .

RUN go build -o server .

# Stage 2: Run
FROM scratch 

WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080

CMD ["./server"]
