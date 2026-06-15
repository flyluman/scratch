FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/scratch ./cmd/server \
    && upx --best --lzma /app/scratch

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1001 appuser

WORKDIR /app

COPY --from=builder /app/scratch /usr/local/bin/scratch
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

USER appuser
ENTRYPOINT ["scratch"]
