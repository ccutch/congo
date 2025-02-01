package congo_chat

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/google/uuid"
)

func (chat *CongoChat) NewMailbox(ownerID, name string, maxSize int) (*Mailbox, error) {
	return chat.NewMailboxWithID(uuid.NewString(), ownerID, name, maxSize)
}

func (chat *CongoChat) NewMailboxWithID(id, ownerID, name string, maxSize int) (*Mailbox, error) {
	m := Mailbox{chat.db.NewModel(id), chat, ownerID, name, maxSize}
	return &m, chat.db.Query(`

		INSERT INTO mailboxes (id, owner_id, name, max_size)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, m.ID, m.OwnerID, m.Name, m.MaxSize).Scan(&m.CreatedAt, &m.UpdatedAt)
}

func (chat *CongoChat) GetMailbox(id string) (*Mailbox, error) {
	m := Mailbox{Model: chat.db.Model(), chat: chat}
	return &m, chat.db.Query(`

		SELECT id, owner_id, name, max_size, created_at, updated_at
		FROM mailboxes
		WHERE id = ?

	`, id).Scan(&m.ID, &m.OwnerID, &m.Name, &m.MaxSize, &m.CreatedAt, &m.UpdatedAt)
}

func (chat *CongoChat) GetMailboxForOwner(ownerID string) (*Mailbox, error) {
	m := Mailbox{Model: chat.db.Model(), chat: chat}
	return &m, chat.db.Query(`

		SELECT id, owner_id, name, max_size, created_at, updated_at
		FROM mailboxes
		WHERE owner_id = ?
		LIMIT 1

	`, ownerID).Scan(&m.ID, &m.OwnerID, &m.Name, &m.MaxSize, &m.CreatedAt, &m.UpdatedAt)
}

func (chat *CongoChat) AllMailboxes() ([]*Mailbox, error) {
	var mailboxes []*Mailbox
	return mailboxes, chat.db.Query(`

		SELECT id, owner_id, name, max_size, created_at, updated_at
		FROM mailboxes

	`).All(func(scan congo.Scanner) error {
		m := Mailbox{Model: chat.db.Model(), chat: chat}
		mailboxes = append(mailboxes, &m)
		return scan(&m.ID, &m.OwnerID, &m.Name, &m.MaxSize, &m.CreatedAt, &m.UpdatedAt)
	})
}

type Mailbox struct {
	congo.Model
	chat    *CongoChat
	OwnerID string
	Name    string
	MaxSize int
}

func (mb *Mailbox) Owner() *congo_auth.Identity {
	mb, err := mb.chat.GetMailbox(mb.OwnerID)
	if err != nil {
		return &congo_auth.Identity{
			Model: mb.chat.db.NewModel(mb.OwnerID),
			Role:  "anon",
			Name:  "unknown",
		}
	}

	id, err := mb.chat.auth.Lookup(mb.OwnerID)
	if err != nil {
		agent, err := mb.chat.GetChatbot(mb.OwnerID)
		if err != nil {
			return &congo_auth.Identity{
				Model: mb.chat.db.NewModel(mb.OwnerID),
				Role:  "anon",
				Name:  "unknown",
			}
		}
		return &congo_auth.Identity{
			Model: mb.chat.db.NewModel(mb.OwnerID),
			Role:  "chatbot",
			Name:  agent.Name,
		}
	}
	return id
}

func (mb *Mailbox) Save() error {
	return mb.chat.db.Query(`

		UPDATE mailboxes
		SET name = ?,
				max_size = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, mb.Name, mb.MaxSize, mb.ID).Scan(&mb.UpdatedAt)
}

func (mb *Mailbox) Delete() error {
	return mb.chat.db.Query(`

		DELETE FROM mailboxes
		WHERE id = ?

	`, mb.ID).Exec()
}
