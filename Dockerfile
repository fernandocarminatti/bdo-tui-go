FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bdo-tui .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bdo-tui .
RUN apk --no-cache add tzdata
CMD ["./bdo-tui"]