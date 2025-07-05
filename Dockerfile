ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /notify-agent ./cmd/agent

FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.title="Notify Agent" \
      org.opencontainers.image.description="Forward Docker/systemd events to Telegram" \
      org.opencontainers.image.url="https://github.com/you/notify-agent" \
      org.opencontainers.image.version="${VERSION:-dev}"

COPY --from=build /notify-agent /notify-agent

USER nonroot
ENTRYPOINT ["/notify-agent"]
