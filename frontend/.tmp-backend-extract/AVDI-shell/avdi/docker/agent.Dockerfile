FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent

FROM alpine:3.20
RUN apk add --no-cache iputils iproute2 net-tools nmap curl bind-tools
WORKDIR /app
COPY --from=builder /agent /agent
COPY diagnostics.yaml /etc/avdi/diagnostics.yaml
CMD ["/agent"]