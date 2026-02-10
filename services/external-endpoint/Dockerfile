FROM golang:1.25.4-alpine AS builder

WORKDIR /build

COPY pkg/ pkg/
COPY services/external-endpoint/ services/external-endpoint/

WORKDIR /build/services/external-endpoint

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/external-endpoint ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/external-endpoint .

EXPOSE 8081

CMD ["./external-endpoint"]
