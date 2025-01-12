package congo_chat

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

//go:embed all:migrations
var migrations embed.FS

//go:embed all:templates
var Templates embed.FS

type CongoChat struct {
	db     *congo.Database
	auth   *congo_auth.CongoAuth
	events map[string][]chan *Message
}

func InitCongoChat(root string, auth *congo_auth.CongoAuth) *CongoChat {
	db := congo.SetupDatabase(root, "chat.db", migrations)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup chat db:", err)
	}
	return &CongoChat{db, auth, map[string][]chan *Message{}}
}

func (chat *CongoChat) Listen(id string) (chan *Message, func()) {
	feed := make(chan *Message)
	chat.events[id] = append(chat.events[id], feed)
	return feed, func() {
		close(feed)
	}
}

func (chat *CongoChat) Notify(m *Message) {
	to := m.ToID
	log.Println("Notifying", to)
	if l, ok := chat.events[to]; ok {
		for i, ch := range l {
			select {
			case ch <- m:
				continue
			default:
				chat.events[to] = append(chat.events[to][:i], chat.events[to][i+1:]...)
				i--
			}
		}
	}
}
