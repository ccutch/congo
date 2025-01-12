package congo_chat

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Attachment struct {
	congo.Model
	chat      *CongoChat
	MessageID string
	Filename  string
	Filetype  string
	Content   []byte
}

func (m *Message) Attachment(filename, filetype string, content []byte) (*Attachment, error) {
	a := Attachment{m.chat.db.NewModel(uuid.NewString()), m.chat, m.ID, filename, filetype, content}
	return &a, m.chat.db.Query(`

		INSERT INTO attachments (id, message_id, name, content_type, content)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, a.ID, a.MessageID, a.Filename, a.Filetype, a.Content).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (m *Message) Attachments() ([]*Attachment, error) {
	var attachments []*Attachment
	return attachments, m.chat.db.Query(`

		SELECT id, message_id, name, content_type, content, created_at, updated_at
		FROM attachments
		WHERE message_id = ?
		ORDER BY created_at DESC

	`, m.ID).All(func(scan congo.Scanner) error {
		a := Attachment{Model: m.chat.db.Model(), chat: m.chat}
		attachments = append(attachments, &a)
		return scan(&a.ID, &a.MessageID, &a.Filename, &a.Filetype, &a.Content, &a.CreatedAt, &a.UpdatedAt)
	})
}

func (a *Attachment) Save() error {
	return a.chat.db.Query(`

		UPDATE attachments
		SET name = ?,
				content_type = ?,
				content = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, a.Filename, a.Filetype, a.Content, a.ID).Scan(&a.UpdatedAt)
}

func (a *Attachment) Delete() error {
	return a.chat.db.Query(`

		DELETE FROM attachments
		WHERE id = ?

	`, a.ID).Exec()
}
