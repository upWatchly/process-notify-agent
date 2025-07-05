package watcher

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerWatcher struct {
	cli  *client.Client
	flt  filters.Args
	host string
}

func NewDockerWatcher(eventFilter string) (*DockerWatcher, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	args := filters.NewArgs(filters.Arg("type", "container"))
	for _, e := range ParseList(eventFilter) {
		args.Add("event", e)
	}
	return &DockerWatcher{cli: cli, flt: args, host: hostname()}, nil
}

func (w *DockerWatcher) Run(ctx context.Context, out chan<- Event) error {
	msgs, errs := w.cli.Events(ctx, events.ListOptions{Filters: w.flt})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return err
		case m := <-msgs:
			exit := parseExit(m)
			img := m.Actor.Attributes["image"]
			if img == "" {
				img = m.From
			}
			act := string(m.Action)
			if act == "die" && m.Actor.Attributes["oomKilled"] == "true" {
				act = "oom-killed"
			}
			if act == "die" && exit != nil && *exit == 0 {
				act = "completed"
			}
			out <- Event{
				Source:     "docker",
				Service:    m.Actor.Attributes["name"],
				Action:     act,
				Image:      img,
				ExitCode:   exit,
				Host:       w.host,
				OccurredAt: time.Unix(m.Time, 0),
			}
		}
	}
}

func parseExit(m events.Message) *int {
	if s, ok := m.Actor.Attributes["exitCode"]; ok {
		if v, err := strconv.Atoi(s); err == nil {
			return &v
		}
	}
	return nil
}

func hostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}
	return "unknown"
}
