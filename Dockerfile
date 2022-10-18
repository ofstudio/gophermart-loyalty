FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /build/gophermart ./cmd/gophermart/main.go

FROM alpine:3.16
COPY --from=builder /build/gophermart /
USER nobody
CMD ["/gophermart"]
