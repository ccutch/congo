package congo_code

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Workspace struct {
	*Service
	repo *Repository
}

func (code *CongoCode) RunWorkspace(name string, port int, repo *Repository, opts ...ServiceOpt) *Workspace {
	opts = append([]ServiceOpt{
		WithImage("codercom/code-server"),
		WithTag("latest"),
		WithPort(port),
		WithEnv("PORT", strconv.Itoa(port)),
		WithVolume(fmt.Sprintf("%s/services/workspace-%s/.config:/home/coder/.config", code.DB.Root, name)),
		WithVolume(fmt.Sprintf("%s/services/workspace-%s/project:/home/coder/project", code.DB.Root, name)),
		WithArgs("--auth", "none"),
	}, opts...)
	service := code.Service("workspace-"+name, opts...)
	return &Workspace{service, repo}
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
		return fmt.Errorf("failed to prepare workspace: %s", output.String())
	}

	if err := w.Service.Start(); err != nil {
		return err
	}

	if _, output, err = w.code.bash(setupWorkspace); err != nil {
		return fmt.Errorf("failed to setup workspace: %s", output.String())
	}

	if w.repo != nil {
		if token, err := w.code.NewAccessToken(time.Now().Add(100_000 * time.Hour)); err == nil {
			output, errput, _ := w.Run(fmt.Sprintf(cloneRepository, token.ID, token.Secret))
			log.Println("cloning", output.String(), errput.String())
		} else {
			log.Println("Failed to create access token: ", err)
		}
	}

	return nil
}

func (w *Workspace) Run(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return w.code.docker("exec", w.Name, "sh", "-c", strings.Join(args, " "))
}
