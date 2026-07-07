FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pharma-platform ./cmd/pharma-platform

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /pharma-platform .
COPY config/ config/
COPY deploy/ deploy/
EXPOSE 8081
CMD ["./pharma-platform"]
