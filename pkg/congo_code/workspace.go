package congo_code

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/pkg/errors"
)

type Workspace struct {
	code *CongoCode
	congo.Model
	*congo_host.Service
	Port   int
	Ready  bool
	RepoID string
}

func (code *CongoCode) RunWorkspace(host *congo_host.CongoHost, name string, port int, repo *Repository, opts ...congo_host.ServiceOpt) (*Workspace, error) {
	repoID := ""
	if repo != nil {
		repoID = repo.ID
	}

	opts = append([]congo_host.ServiceOpt{
		congo_host.WithImage("codercom/code-server"),
		congo_host.WithTag("latest"),
		congo_host.WithPort(port),
		congo_host.WithEnv("PORT", strconv.Itoa(port)),
		congo_host.WithVolume(fmt.Sprintf("%s/services/workspace-%s/.config:/home/coder/.config", code.db.Root, name)),
		congo_host.WithVolume(fmt.Sprintf("%s/services/workspace-%s/project:/home/coder/project", code.db.Root, name)),
		congo_host.WithArgs("--auth", "none"),
	}, opts...)

	id := fmt.Sprintf("workspace-%s", name)
	w := Workspace{code, code.db.NewModel(id), host.Local().Service(id, opts...), port, false, repoID}
	return &w, code.db.Query(`
	
		INSERT INTO workspaces (id, name, port, image, tag, ready, repo_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at

	`, w.ID, w.Name, w.Port, w.Image, w.Tag, w.Ready, w.RepoID).Scan(&w.CreatedAt, &w.UpdatedAt)
}

func (code *CongoCode) GetWorkspace(id string) (*Workspace, error) {
	w := Workspace{code: code, Model: code.db.Model(), Service: &congo_host.Service{}}
	return &w, code.db.Query(`
	
		SELECT id, name, port, image, tag, ready, repo_id, created_at, updated_at
		FROM workspaces
		WHERE id = ?

	`, id).Scan(&w.ID, &w.Name, &w.Port, &w.Image, &w.Tag, &w.Ready, &w.RepoID, &w.CreatedAt, &w.UpdatedAt)
}

func (code *CongoCode) AllWorkspaces() ([]*Workspace, error) {
	var workspaces []*Workspace
	return workspaces, code.db.Query(`
	
		SELECT id, name, port, image, tag, ready, repo_id, created_at, updated_at
		FROM workspaces
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		w := Workspace{code: code, Model: code.db.Model()}
		workspaces = append(workspaces, &w)
		return scan(&w.ID, &w.Name, &w.Port, &w.Image, &w.Tag, &w.Ready, &w.RepoID, &w.CreatedAt, &w.UpdatedAt)
	})
}

func (code *CongoCode) Count() (count int) {
	code.db.Query(`SELECT count(*) FROM workspaces`).Scan(&count)
	return count
}

//go:embed resources/workspace/prepare-workspace.sh
var prepareWorkspace string

//go:embed resources/workspace/setup-workspace.sh
var setupWorkspace string

//go:embed resources/workspace/clone-repository.sh
var cloneRepository string

func (w *Workspace) Start() error {
	if w.Running() {
		log.Printf("Workspace %s already running", w.Name)
		return nil
	}

	var stdout bytes.Buffer
	host := w.Host.Local()
	if err := host.Run(fmt.Sprintf(prepareWorkspace, w.Name, w.code.db.Root)); err != nil {
		return errors.Wrap(err, "failed to prepare workspace")
	}
	if err := w.Service.Start(); err != nil {
		return errors.Wrap(err, "failed to start workspace")
	}

	stdout.Reset()
	if err := host.Run(setupWorkspace); err != nil {
		return errors.Wrap(err, "failed to setup workspace: "+stdout.String())
	}

	if repo, err := w.Repo(); repo != nil && err == nil {
		if token, err := w.code.NewAccessToken(time.Now().Add(100_000 * time.Hour)); err == nil {
			w.Run(fmt.Sprintf(cloneRepository, token.ID, token.Secret))
		}
	}

	w.Ready = true
	return w.Save()
}

func (w *Workspace) Repo() (*Repository, error) {
	if w.RepoID == "" {
		return nil, nil
	}
	return w.code.GetRepository(w.RepoID)
}

func (w *Workspace) Run(args ...string) (stdout bytes.Buffer, err error) {
	s := w.Host.Local()
	s.SetStdout(&stdout)
	return stdout, s.Run("docker", "exec", w.Name, "sh", "-c", strings.Join(args, " "))
}

//go:embed resources/workspace/create-congo-app.sh
var createCongoApp string

func (w *Workspace) CreateCongoApp(name, template string) error {
	return w.Host.Local().Run(fmt.Sprintf(createCongoApp, name, template))
}

func (w *Workspace) Save() error {
	return w.DB.Query(`
	
		UPDATE workspaces
		SET ready = ?,
				name = ?,
				port = ?,
				image = ?,
				tag = ?,
				repo_id = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, w.Ready, w.Name, w.Port, w.Image, w.Tag, w.RepoID, w.ID).Scan(&w.UpdatedAt)
}
