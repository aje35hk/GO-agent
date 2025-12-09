FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o agent main.go

FROM alpine:latest
RUN apk add --no-cache bash
WORKDIR /root/
COPY --from=builder /app/agent .

CMD ["./agent"]
