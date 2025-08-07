ARG GO_VERSION=1.24.5

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0  go build -ldflags '-s -w' -o server .

FROM gcr.io/distroless/static AS production

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]