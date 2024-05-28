FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ./cmd/server.go

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/server .

RUN chmod +x ./server

CMD ["./server"]
