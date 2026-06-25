FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /out/cli-login ./cmd/cli

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app && mkdir -p /app/data && chown -R app:app /app

COPY --from=build /out/cli-login /app/cli-login

USER app

ENV DB_PATH=/app/data/app.db
ENV SESSION_TIMEOUT_MINUTES=30
ENV MAX_FAILED_ATTEMPTS=5
ENV LOCKOUT_MINUTES=15

CMD ["/app/cli-login"]
