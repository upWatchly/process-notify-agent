package watcher

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Event struct {
	Source     string    // docker | systemd | ...
	Image      string    // container image, unit name, etc.
	Service    string    // container name, unit name, etc.
	Action     string    // start | die | restart | health_status:unhealthy | failed
	ExitCode   *int      // nil if N/A
	Host       string    // hostname ‑ filled by watcher
	OccurredAt time.Time // timestamp of event
}

type Watcher interface {
	Run(ctx context.Context, out chan<- Event) error
}

func ParseList(csv string) []string {
	var out []string
	for _, s := range strings.Split(csv, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func Factory(kind, filter string) (Watcher, error) {
	switch kind {
	case "docker":
		return NewDockerWatcher(filter)
	default:
		return nil, fmt.Errorf("unknown source %s", kind)
	}
}
