package congo_code

import (
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type AccessToken struct {
	congo.Model
	Secret  string
	Expires time.Time
}

func (code *CongoCode) NewAccessToken(expires time.Time) (*AccessToken, error) {
	t := AccessToken{Model: code.db.NewModel(uuid.NewString())}
	return &t, code.db.Query(`
	
		INSERT INTO access_tokens (id, secret, expires_at)
		VALUES (?, ?, ?)
		RETURNING secret, expires_at, created_at, updated_at

	`, t.ID, uuid.NewString(), expires).Scan(&t.Secret, &t.Expires, &t.CreatedAt, &t.UpdatedAt)
}

func (code *CongoCode) GetAccessToken(id, secret string) (*AccessToken, error) {
	t := AccessToken{Model: code.db.NewModel(id)}
	return &t, code.db.Query(`
	
		SELECT secret, expires_at, created_at, updated_at
		FROM access_tokens
		WHERE id = ? AND secret = ? AND expires_at > CURRENT_TIMESTAMP
	
	`, id, secret).Scan(&t.Secret, &t.Expires, &t.CreatedAt, &t.UpdatedAt)
}
