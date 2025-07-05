package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"process-notify-agent/internal/sink"
	"process-notify-agent/internal/watcher"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	chatID := os.Getenv("CHAT_ID")
	var (
		sources   string
		eventFilt string
	)
	flag.StringVar(&sources, "sources", "docker", "comma‑separated list: docker,systemd,supervisor,pm2")
	flag.StringVar(&eventFilt, "filter", "start,die,restart,health_status", "comma‑separated docker events to forward")
	flag.Parse()

	if botToken == "" || chatID == "" {
		log.Fatal("BOT_TOKEN and CHAT_ID must be set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	events := make(chan watcher.Event, 100)

	for _, s := range watcher.ParseList(sources) {
		w, err := watcher.Factory(s, eventFilt)
		if err != nil {
			log.Warnf("ignore source %s: %v", s, err)
			continue
		}
		go func(w watcher.Watcher) {
			if err := w.Run(ctx, events); err != nil {
				log.Errorf("watcher %T stopped: %v", w, err)
			}
		}(w)
	}

	ts := sink.NewTelegramSink(botToken, chatID)

	for {
		select {
		case <-ctx.Done():
			log.Info("shutdown")
			return
		case ev := <-events:
			ts.Send(ev)
		}
	}
}
