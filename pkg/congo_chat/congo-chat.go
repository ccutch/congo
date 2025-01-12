package congo_chat

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

//go:embed all:migrations
var migrations embed.FS

type CongoChat struct {
	db    *congo.Database
	auth  *congo_auth.CongoAuth
	feeds map[string][]*Listener
}

func InitCongoChat(root string, auth *congo_auth.CongoAuth) *CongoChat {
	db := congo.SetupDatabase(root, "chat.db", migrations)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup chat db:", err)
	}
	return &CongoChat{db, auth, map[string][]*Listener{}}
}

type Listener struct {
	Messages chan *Message
	Closed   bool
	Close    func()
}

func (chat *CongoChat) Listen(id string) (*Listener, func()) {
	feed := make(chan *Message)
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
	if f, ok := chat.feeds[m.FromID]; ok {
		for _, l := range f {
			if !l.Closed {
				l.Messages <- m
			}
		}
	}
}
