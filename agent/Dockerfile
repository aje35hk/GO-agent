FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o agent main.go

FROM alpine:latest
RUN apk add --no-cache bash curl
WORKDIR /root/
COPY --from=builder /app/agent .

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

CMD ["./agent"]
