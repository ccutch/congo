package congo_chat

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Message struct {
	congo.Model
	chat    *CongoChat
	ToID    string
	FromID  string
	Content string
}

func (chat *CongoChat) GetMessage(id string) (*Message, error) {
	m := Message{Model: chat.db.Model(), chat: chat}
	return &m, chat.db.Query(`

		SELECT id, to_mailbox, from_mailbox, content, created_at, updated_at
		FROM messages
		WHERE id = ?

	`, id).Scan(&m.ID, &m.ToID, &m.FromID, &m.Content, &m.CreatedAt, &m.UpdatedAt)
}

func (mb *Mailbox) Send(to, content string) (*Message, error) {
	m := Message{mb.chat.db.NewModel(uuid.NewString()), mb.chat, to, mb.ID, content}
	err := mb.chat.db.Query(`

		INSERT INTO messages (id, to_mailbox, from_mailbox, content)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, m.ID, m.ToID, m.FromID, m.Content).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	mb.chat.Notify(&m)
	return &m, err
}

func (mb *Mailbox) Contacts() ([]*Mailbox, error) {
	var contacts []*Mailbox
	return contacts, mb.chat.db.Query(`

		SELECT id, owner_id, name, max_size, created_at, updated_at
		FROM mailboxes
		WHERE id IN (
			SELECT from_mailbox
			FROM messages
			WHERE to_mailbox = $1
			  AND id != $1
		)

	`, mb.ID).All(func(scan congo.Scanner) error {
		m := Mailbox{Model: mb.chat.db.Model(), chat: mb.chat}
		contacts = append(contacts, &m)
		return scan(&m.ID, &m.OwnerID, &m.Name, &m.MaxSize, &m.CreatedAt, &m.UpdatedAt)
	})
}

func (mb *Mailbox) CountMessages() (count int) {
	mb.chat.db.Query(`

		SELECT count(*)
		FROM messages
		WHERE to_mailbox = ?

	`, mb.ID).Scan(&count)
	return count
}

func (mb *Mailbox) Messages(from string) ([]*Message, error) {
	var messages []*Message
	return messages, mb.chat.db.Query(`

		SELECT id, to_mailbox, from_mailbox, content, created_at, updated_at
		FROM messages
		WHERE (to_mailbox = $1 AND from_mailbox = $2)
			 OR (to_mailbox = $2 AND from_mailbox = $1)
		ORDER BY created_at DESC

	`, mb.ID, from).All(func(scan congo.Scanner) error {
		m := Message{Model: mb.chat.db.Model(), chat: mb.chat}
		messages = append(messages, &m)
		return scan(&m.ID, &m.ToID, &m.FromID, &m.Content, &m.CreatedAt, &m.UpdatedAt)
	})
}

func (m *Message) Save() error {
	return m.chat.db.Query(`

		UPDATE messages
		SET to_mailbox = ?,
				from_mailbox = ?,
				content = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, m.ToID, m.FromID, m.Content, m.ID).Scan(&m.UpdatedAt)
}

func (m *Message) Delete() error {
	return m.chat.db.Query(`

		DELETE FROM messages
		WHERE id = ?

	`, m.ID).Exec()
}
