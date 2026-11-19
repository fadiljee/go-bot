# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bot-istimewa ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=build /out/bot-istimewa /app/bot-istimewa
COPY locales /app/locales
COPY configs /app/configs
COPY StikerChatbot /app/StikerChatbot

RUN mkdir -p /app/storage && chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

ENTRYPOINT ["/app/bot-istimewa"]
