# Process Notify Agent

Lightweight Go service that listens for container and service events and sends them to Telegram.

* Watches Docker containers, systemd units, and (optionally) Supervisor / PM2
* Ships as a single static binary **or** a 10 MB distroless image
* Tags each message for easy filtering (`#host_…`, `#img_…`, `#ctr_…`)
* Works on Linux (amd64 / arm64) **and** macOS (`DOCKER_HOST=unix:///path/to/docker.sock`)

---

## Quick start — Docker Compose

    version: "3.9"

    services:
      notify-agent:
        image: ghcr.io/you/notify-agent:0.3.0
        restart: unless-stopped

        environment:
          - BOT_TOKEN=123456:ABCDEF…
          - CHAT_ID=-100987654321          # channel or user ID
          - SOURCES=docker,systemd         # watchers: docker, systemd, supervisor, pm2
          - FILTER=start,die,restart,oom,health_status
          - DOCKER_HOST=unix:///docker.sock   # change if your socket lives elsewhere

        volumes:
          - /var/run/docker.sock:/docker.sock:ro
          - /run/dbus/system_bus_socket:/run/dbus/system_bus_socket:ro   # systemd events

Bring it up:

    docker compose up -d

Kill a test container and watch formatted messages appear in Telegram.

---

## Running the binary

    wget https://github.com/you/notify-agent/releases/download/v0.3.0/notify-agent_0.3.0_linux_amd64
    chmod +x notify-agent
    export BOT_TOKEN=… CHAT_ID=…
    ./notify-agent --sources=docker --filter=start,die,restart

A systemd unit example is in `packaging/notify-agent.service`.

---

## Configuration

| Variable / Flag                 | Default                | Description                               |
|---------------------------------|------------------------|-------------------------------------------|
| `BOT_TOKEN`                     | —                      | Telegram bot token                        |
| `CHAT_ID`                       | —                      | Chat / channel ID (negative for channels) |
| `SOURCES`                       | `docker`               | `docker`, `systemd`, `supervisor`, `pm2`  |
| `FILTER`                        | `start,die,restart`    | Comma-separated list of events            |
| `DOCKER_HOST` / `--docker-sock` | `/var/run/docker.sock` | Path or TCP URI to the Docker socket      |

---

## Build from source

    git clone https://github.com/you/notify-agent.git
    cd notify-agent
    CGO_ENABLED=0 go build -ldflags "-s -w" -o notify-agent ./cmd/agent

Build the image:

    docker build -t notify-agent:dev .

---

## Contributing

1. Fork the repo
2. Create a branch, run `go fmt ./...`
3. Ensure `go test ./...` is green
4. Open a PR

Real-world feedback is welcome — open an issue with log snippets.

---

## License

MIT — see `LICENSE`.
