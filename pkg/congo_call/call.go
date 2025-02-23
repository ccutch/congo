package congo_call

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Call struct {
	congo.Model
	Name   string
	peers  []*Peer
	Offer  *CallOffer
	Answer *CallAnswer
}

func (c *CongoCall) NewRoom(name string) (*Call, error) {
	r := Call{Model: c.db.NewModel(uuid.NewString()), Name: name, peers: []*Peer{}}
	return &r, c.db.Query(`

		INSERT INTO calls (id, name)
		VALUES (?, ?)
		RETURNING created_at, updated_at

	`, r.ID, r.Name).Scan(&r.CreatedAt, &r.UpdatedAt)
}

func (c *CongoCall) GetRoom(id string) (*Call, error) {
	r := Call{Model: c.db.Model(), peers: []*Peer{}}
	return &r, c.db.Query(`

		SELECT id, name, created_at, updated_at
		FROM calls
		WHERE id = ?

	`, id).Scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt)
}

func (c *CongoCall) Calls() (calls []*Call, err error) {
	return calls, c.db.Query(`

		SELECT id, name, created_at, updated_at
		FROM calls
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		r := Call{Model: c.db.Model(), peers: []*Peer{}}
		err = scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt)
		calls = append(calls, &r)
		return err
	})
}

func (r *Call) Save() error {
	return r.DB.Query(`

		UPDATE calls
		SET name = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, r.Name, r.ID).Scan(&r.UpdatedAt)
}

func (r *Call) Delete() error {
	return r.DB.Query(`

		DELETE FROM calls
		WHERE id = ?

	`, r.ID).Exec()
}
