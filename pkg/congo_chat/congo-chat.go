package congo_chat

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
)

//go:embed all:migrations
var migrations embed.FS

type CongoChat struct {
	db    *congo.Database
	auth  *congo_auth.CongoAuth
	feeds map[string][]*Listener
	host  string
	Model string
}

func InitCongoChat(root string, opts ...CongoChatOptions) *CongoChat {
	db := congo.SetupDatabase(root, "chat.db", migrations)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup chat db:", err)
	}
	chat := CongoChat{db, nil, map[string][]*Listener{}, "http://localhost:11434", ""}
	for _, opt := range opts {
		opt(&chat)
	}
	return &chat
}

type CongoChatOptions func(*CongoChat)

func WithAuth(auth *congo_auth.CongoAuth) CongoChatOptions {
	return func(chat *CongoChat) {
		chat.auth = auth
	}
}

func WithModel(name string) CongoChatOptions {
	log.Println("Going to start ollama model", name)
	return func(chat *CongoChat) {
		log.Println("Loading ollama model", name)
		chat.Model = name
		host := congo_host.InitCongoHost(chat.db.Root, nil)
		service := host.Local().Service("ollama",
			congo_host.WithImage("ollama/ollama"),
			congo_host.WithTag("latest"),
			congo_host.WithPort(11434),
			congo_host.WithVolume(fmt.Sprintf("%s/services/ollama:/root/.ollama", host.DB.Root)))
		go func() {
			service.Start()
			time.Sleep(10 * time.Second)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(map[string]any{
				"model": name,
			})

			log.Printf("Calling %s/api/pull", chat.host)
			resp, err := http.Post(chat.host+"/api/pull", "application/json", &body)
			if err != nil {
				log.Fatal("Failed to pull ollama model:", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Fatal("Failed to pull ollama model:", resp.Status)
			}

			var buf bytes.Buffer
			_, err = io.Copy(&buf, resp.Body)
			log.Println(buf.String(), err)
			log.Printf("Pulled ollama model %s", name)
		}()
	}
}

type Listener struct {
	Messages chan *Message
	Closed   bool
	Close    func()
}

func (chat *CongoChat) Listen(id string) (*Listener, func()) {
	feed := make(chan *Message, 100)
	listener := &Listener{
		Messages: feed,
		Closed:   false,
		Close:    func() { close(feed) },
	}
	chat.feeds[id] = append(chat.feeds[id], listener)
	return listener, func() {
		listener.Closed = true
		close(feed)
	}
}

func (chat *CongoChat) Notify(m *Message) {
	if f, ok := chat.feeds[m.ToID]; ok {
		for _, l := range f {
			if !l.Closed {
				l.Messages <- m
				continue
			}
		}
	}
	if m.FromID == m.ToID {
		return
	}
	if f, ok := chat.feeds[m.FromID]; ok {
		for _, l := range f {
			if !l.Closed {
				l.Messages <- m
			}
		}
	}
}
