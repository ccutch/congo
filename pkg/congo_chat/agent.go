package congo_chat

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Agent struct {
	congo.Model
	chat         *CongoChat
	Name         string
	ModelName    string
	SystemPrompt string
}

func (chat *CongoChat) NewAgent(name, model, systemPrompt string) (*Agent, error) {
	a := Agent{chat.db.NewModel(uuid.NewString()), chat, name, model, systemPrompt}
	return &a, chat.db.Query(`

		INSERT INTO agents (id, name, model, system_prompt)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, a.ID, a.Name, a.Model, a.SystemPrompt).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (chat *CongoChat) GetAgent(id string) (*Agent, error) {
	a := Agent{Model: chat.db.Model()}
	return &a, chat.db.Query(`

		SELECT id, name, model, system_prompt, created_at, updated_at
		FROM agents
		WHERE id = ?

	`, id).Scan(&a.ID, &a.Name, &a.Model, &a.SystemPrompt, &a.CreatedAt, &a.UpdatedAt)
}

func (chat *CongoChat) CountAgents() (count int) {
	chat.db.Query(`SELECT count(*) FROM agents`).Scan(&count)
	return count
}

func (chat *CongoChat) AllAgents() ([]*Agent, error) {
	var agents []*Agent
	return agents, chat.db.Query(`

		SELECT id, name, model, system_prompt, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		a := Agent{Model: chat.db.Model(), chat: chat}
		agents = append(agents, &a)
		return scan(&a.ID, &a.Name, &a.Model, &a.SystemPrompt, &a.CreatedAt, &a.UpdatedAt)
	})
}

func (chat *CongoChat) SearchAgents(query string) (agents []*Agent, err error) {
	return agents, chat.db.Query(`

		SELECT id, name, model, system_prompt, created_at, updated_at
		FROM agents
		WHERE id LIKE $1 OR name LIKE $1 OR model LIKE $1

	`, "%"+query+"%").All(func(scan congo.Scanner) error {
		a := Agent{Model: chat.db.Model(), chat: chat}
		agents = append(agents, &a)
		return scan(&a.ID, &a.Name, &a.Model, &a.SystemPrompt, &a.CreatedAt, &a.UpdatedAt)
	})
}

func (a *Agent) Save() error {
	return a.DB.Query(`

		UPDATE agents
		SET name = ?,
				model = ?,
				system_prompt = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, a.Name, a.Model, a.SystemPrompt, a.ID).Scan(&a.UpdatedAt)
}

func (a *Agent) Delete() error {
	return a.DB.Query(`

		DELETE FROM agents
		WHERE id = ?

	`, a.ID).Exec()
}
