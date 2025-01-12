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

	`, ownerID).Scan(&m.ID, &m.OwnerID, &m.Name, &m.MaxSize, &m.CreatedAt, &m.UpdatedAt)
}

type Mailbox struct {
	congo.Model
	chat    *CongoChat
	OwnerID string
	Name    string
	MaxSize int
}

func (mb *Mailbox) Owner() (*congo_auth.Identity, error) {
	i, err := mb.chat.auth.Lookup(mb.OwnerID)
	if err != nil {
		return nil, err
	}
	return i, nil
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
