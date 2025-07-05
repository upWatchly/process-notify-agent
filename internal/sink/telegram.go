package sink

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"process-notify-agent/internal/watcher"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var emoji = map[string]string{
	"start":                   "🟢",
	"restart":                 "🔄",
	"die":                     "❌",
	"oom-killed":              "💀",
	"completed":               "✅",
	"health_status:unhealthy": "⚠️",
}

var mdEsc = regexp.MustCompile(`([_\*\[\]()~` + "`" + `])`)

func escHTML(s string) string {
	return html.EscapeString(s)
}

func tag(key, val string) string {
	clean := strings.NewReplacer(" ", "_", "-", "_").Replace(val)
	return fmt.Sprintf("#%s_%s", key, clean)
}

type TelegramSink struct {
	token string
	chat  string
	cli   *http.Client
}

func NewTelegramSink(token, chat string) *TelegramSink {
	return &TelegramSink{token: token, chat: chat, cli: &http.Client{Timeout: 5 * time.Second}}
}

func (t *TelegramSink) Send(ev watcher.Event) {
	ico := emoji[ev.Action]
	if ico == "" {
		ico = "ℹ️"
	}

	header := fmt.Sprintf("%s <b>%s</b>", ico, escHTML(ev.Service))

	body := fmt.Sprintf(
		"<b>Host:</b> <code>%s</code>\n"+
			"<b>Image:</b> <code>%s</code>\n"+
			"<b>Status:</b> <b>%s</b>\n"+
			"<b>Time:</b> %s",
		escHTML(ev.Host),
		escHTML(ev.Image),
		strings.ReplaceAll(strings.Title(ev.Action), "_", " "),
		ev.OccurredAt.Format("2006-01-02 15:04:05"),
	)
	tags := strings.Join([]string{
		tag("host", ev.Host),
		tag("img", strings.Split(ev.Image, ":")[0]),
		tag("ctr", ev.Service),
	}, "\n")

	text := fmt.Sprintf("%s\n\n%s\n\n%s", header, body, tags)

	data := url.Values{
		"chat_id":    {t.chat},
		"text":       {text},
		"parse_mode": {"HTML"},
	}
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	resp, err := t.cli.PostForm(api, data)
	if err != nil {
		log.Error(err)
		return
	}
	_ = resp.Body.Close()
}
