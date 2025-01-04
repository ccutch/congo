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
	"github.com/pkg/errors"
)

type Workspace struct {
	congo.Model
	*Service
	Port   int
	Ready  bool
	RepoID string
}

func (code *CongoCode) RunWorkspace(name string, port int, repo *Repository, opts ...ServiceOpt) (*Workspace, error) {
	repoID := ""
	if repo != nil {
		repoID = repo.ID
	}

	opts = append([]ServiceOpt{
		WithImage("codercom/code-server"),
		WithTag("latest"),
		WithPort(port),
		WithEnv("PORT", strconv.Itoa(port)),
		WithVolume(fmt.Sprintf("%s/services/workspace-%s/.config:/home/coder/.config", code.DB.Root, name)),
		WithVolume(fmt.Sprintf("%s/services/workspace-%s/project:/home/coder/project", code.DB.Root, name)),
		WithArgs("--auth", "none"),
	}, opts...)

	id := fmt.Sprintf("workspace-%s", name)
	w := Workspace{code.DB.NewModel(id), code.Service(id, opts...), port, false, repoID}
	return &w, code.DB.Query(`
	
		INSERT INTO workspaces (id, name, port, image, tag, ready, repo_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, w.ID, w.Name, w.Port, w.Image, w.Tag, w.Ready, w.RepoID).Scan(&w.CreatedAt, &w.UpdatedAt)
}

func (code *CongoCode) GetWorkspace(id string) (*Workspace, error) {
	w := Workspace{Model: code.DB.Model()}
	return &w, code.DB.Query(`
	
		SELECT id, name, port, image, tag, ready, repo_id, created_at, updated_at
		FROM workspaces
		WHERE id = ?

	`, id).Scan(&w.ID, &w.Name, &w.Port, &w.Image, &w.Tag, &w.Ready, &w.RepoID, &w.CreatedAt, &w.UpdatedAt)
}

func (code *CongoCode) AllWorkspaces() ([]*Workspace, error) {
	var workspaces []*Workspace
	return workspaces, code.DB.Query(`
	
		SELECT id, name, port, image, tag, ready, repo_id, created_at, updated_at
		FROM workspaces
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		w := Workspace{Model: code.DB.Model()}
		workspaces = append(workspaces, &w)
		return scan(&w.ID, &w.Name, &w.Port, &w.Image, &w.Tag, &w.Ready, &w.RepoID, &w.CreatedAt, &w.UpdatedAt)
	})
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

	_, output, err := w.code.bash(fmt.Sprintf(prepareWorkspace, w.Name, w.code.DB.Root))
	if err != nil {
		return errors.Wrap(err, "failed to prepare workspace")
	}
	if err := w.Service.Start(); err != nil {
		return errors.Wrap(err, "failed to start workspace")
	}
	if _, output, err = w.Run(setupWorkspace); err != nil {
		return errors.Wrap(err, "failed to setup workspace: "+output.String())
	}

	if repo, err := w.Repo(); repo != nil && err == nil {
		if token, err := w.code.NewAccessToken(time.Now().Add(100_000 * time.Hour)); err == nil {
			output, errput, _ := w.Run(fmt.Sprintf(cloneRepository, token.ID, token.Secret))
			log.Println("cloning", output.String(), errput.String())
		} else {
			log.Println("Failed to create access token: ", err)
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

func (w *Workspace) Run(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return w.code.docker("exec", w.Name, "sh", "-c", strings.Join(args, " "))
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
