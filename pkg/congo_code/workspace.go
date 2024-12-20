package congo_code

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

//go:embed resources/setup-workspace.sh
var setupWorkspace string

//go:embed resources/start-workspace.sh
var startWorkspace string

//go:embed resources/clone-repository.sh
var cloneRepository string

type Workspace struct {
	Name string
	Port int
	code *CongoCode
	repo *Repository
	*httputil.ReverseProxy
}

type WorkspaceOpt func(*Workspace) error

func (code *CongoCode) Workspace(name string, opts ...WorkspaceOpt) (*Workspace, error) {
	w := Workspace{name, 7000, code, nil, nil}
	for _, opt := range opts {
		if err := opt(&w); err != nil {
			return nil, err
		}
	}
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", w.Port))
	if err != nil {
		return nil, err
	}
	w.ReverseProxy = httputil.NewSingleHostReverseProxy(url)
	return &w, nil
}

func (w *Workspace) Running() bool {
	stdout, _, err := w.code.docker("inspect", "-f", "{{.State.Status}}", w.Name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
}

func (w *Workspace) Start() error {
	if w.Running() {
		log.Printf("Workspace %s already running", w.Name)
		return nil
	}
	_, output, err := w.code.bash(fmt.Sprintf(setupWorkspace, w.Name, w.code.app.DB.Root))
	if err != nil {
		return fmt.Errorf("failed to setup workspace: %s", output.String())
	}
	time.Sleep(time.Second)
	_, output, err = w.code.bash(fmt.Sprintf(startWorkspace, w.Name, w.code.app.DB.Root, w.Port))
	if err != nil {
		return errors.Join(errors.New("failed to start workspace"), err)
	}
	if w.repo != nil {
		output, errput, _ := w.Run(cloneRepository)
		log.Println("cloning", output.String(), errput.String())
	} else {
		log.Println("No repo detected")
	}
	return nil
}

func (w *Workspace) Run(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return w.code.docker("exec", w.Name, "sh", "-c", strings.Join(args, " "))
}

func (w *Workspace) Stop() error {
	if !w.Running() {
		log.Printf("Workspace %s is not running", w.Name)
		return nil
	}
	if _, _, err := w.code.docker("stop", w.Name); err != nil {
		return fmt.Errorf("failed to stop workspace %s: %w", w.Name, err)
	}
	if _, _, err := w.code.docker("rm", w.Name); err != nil {
		return fmt.Errorf("failed to remove workspace %s: %w", w.Name, err)
	}
	return nil
}

func WithPort(port int) WorkspaceOpt {
	return func(w *Workspace) error {
		w.Port = port
		return nil
	}
}

func WithRepo(repo *Repository) WorkspaceOpt {
	return func(w *Workspace) error {
		w.repo = repo
		return nil
	}
}
