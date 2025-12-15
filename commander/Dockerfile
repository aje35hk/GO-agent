FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o controller main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/controller .
COPY --from=builder /app/index.html .

EXPOSE 8080
CMD ["./controller"]
