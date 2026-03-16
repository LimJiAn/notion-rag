FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /bin/notion-rag ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/notion-rag /app/notion-rag

ENV HTTP_ADDR=:8080
ENV DATA_DIR=/app/data

EXPOSE 8080

CMD ["/app/notion-rag"]
