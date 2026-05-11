# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24.4
ARG ALPINE_VERSION=3.21

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src/server
COPY server/go.mod ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ticketmet ./main

FROM alpine:${ALPINE_VERSION}

RUN addgroup -S ticketmet && adduser -S -G ticketmet ticketmet

WORKDIR /app
COPY --from=build /out/ticketmet /app/ticketmet
COPY client /app/client

ENV PORT=8080
ENV CLIENT_DIR=/app/client

EXPOSE 8080

USER ticketmet

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1

ENTRYPOINT ["/app/ticketmet"]
