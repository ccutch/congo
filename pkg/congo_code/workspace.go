package congo_code

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
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
	code *CongoCode
	repo *Repository
	name string
	port int
}

func (code *CongoCode) Workspace(name string, opts ...WorkspaceOpt) (*Workspace, error) {
	w := Workspace{code, nil, name, 8081}
	for _, opt := range opts {
		if err := opt(&w); err != nil {
			return nil, err
		}
	}
	return &w, nil
}

func (w *Workspace) Running() bool {
	stdout, _, err := w.code.docker("inspect", "-f", "{{.State.Status}}", w.name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
}

func (w *Workspace) Start() error {
	if w.Running() {
		log.Printf("Workspace %s already running", w.name)
		return nil
	}
	_, _, err := w.code.bash(fmt.Sprintf(setupWorkspace, w.name, w.code.app.DB.Root))
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	_, _, err = w.code.bash(fmt.Sprintf(startWorkspace, w.name, w.code.app.DB.Root, w.port))
	if err != nil {
		return err
	}
	if w.repo != nil {
		output, errput, err := w.Run(cloneRepository)
		log.Println("cloning", output.String(), errput.String())
		return err
	} else {
		log.Println("No repo detected")
	}
	return nil
}

func (w *Workspace) Run(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return w.code.docker("exec", w.name, "sh", "-c", strings.Join(args, " "))
}

func (w *Workspace) Server() http.Handler {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", w.port))
	if err != nil {
		log.Fatalf("Failed to forward to: localhost:%d", w.port)
	}
	return httputil.NewSingleHostReverseProxy(url)
}

type WorkspaceOpt func(*Workspace) error

func WithPort(port int) WorkspaceOpt {
	return func(w *Workspace) error {
		w.port = port
		return nil
	}
}

func WithRepo(repo *Repository) WorkspaceOpt {
	return func(w *Workspace) error {
		w.repo = repo
		return nil
	}
}
