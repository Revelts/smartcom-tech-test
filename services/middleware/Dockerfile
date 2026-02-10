FROM golang:1.25.4-alpine AS builder

WORKDIR /build

COPY pkg/ pkg/
COPY services/middleware/ services/middleware/

WORKDIR /build/services/middleware

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/middleware ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/middleware .

EXPOSE 8080

CMD ["./middleware"]
