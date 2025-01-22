FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM scratch AS production

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]
