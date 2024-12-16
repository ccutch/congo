package congo_code

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed resources/prep-workspace.sh
var prepWorkspace string

//go:embed resources/start-workspace.sh
var startWorkspace string

func StartWorkspace(db *congo.Database) error {
	_, _, err := Bash(fmt.Sprintf(prepWorkspace, db.Root))
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	_, _, err = Bash(startWorkspace)
	return err
}

func GetWorkspace(db *congo.Database) error {
	stdout, stderr, err := Docker("inspect", "-f", "{{.State.Status}}", "workspace")
	if err != nil {
		log.Println(stderr.String())
		return fmt.Errorf("workspace not found: %v", err)
	}
	status := strings.TrimSpace(string(stdout.String()))
	if status != "running" {
		return fmt.Errorf("workspace is not running, current status: %s", status)
	}
	return nil
}

func WorkspaceRun(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return Docker("exec", "workspace", "sh", "-c", strings.Join(args, " "))
}

func Run(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

func Bash(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return Run(append([]string{"bash", "-c"}, args...)...)
}

func Docker(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return Run(append([]string{"docker"}, args...)...)
}
