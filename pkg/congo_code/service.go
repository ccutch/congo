package congo_code

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Service struct {
	code  *CongoCode
	Name  string
	Port  int
	image string
	tag   string
	envs  []string
	vols  []string
	args  []string
}

func (code *CongoCode) Service(name string, opts ...ServiceOpt) *Service {
	s := &Service{code, name, 0, "", "latest", []string{}, []string{}, []string{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type ServiceOpt func(*Service)

func WithImage(image string) ServiceOpt {
	return func(s *Service) { s.image = image }
}

func WithTag(tag string) ServiceOpt {
	return func(s *Service) { s.tag = tag }
}

func WithPort(port int) ServiceOpt {
	return func(s *Service) { s.Port = port }
}

func WithEnv(name string, value any) ServiceOpt {
	env := fmt.Sprintf("%s=%s", name, value)
	return func(s *Service) { s.envs = append(s.envs, env) }
}

func WithVolume(volume string) ServiceOpt {
	return func(s *Service) { s.vols = append(s.vols, volume) }
}

func WithArgs(args ...string) ServiceOpt {
	return func(s *Service) { s.args = args }
}

func (s *Service) Running() bool {
	stdout, _, err := s.code.docker("inspect", "-f", "{{.State.Status}}", s.Name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
}

//go:embed resources/service/start-service.sh
var startService string

func (s *Service) Start() error {
	if s.Running() {
		return nil
	}

	if s.image == "" {
		return errors.New("missing image")
	}

	if err := s.setupService(); err != nil {
		return errors.Wrap(err, "failed to setup service")
	}

	envs := strings.Join(s.envs, " -e ")
	if envs != "" {
		envs = "-e " + envs
	}

	volumes := strings.Join(s.vols, " -v ")
	if volumes != "" {
		volumes = "-v " + volumes
	}

	args := strings.Join(s.args, " ")
	_, output, err := s.code.bash(fmt.Sprintf(startService, s.Name, s.Port, envs, volumes, s.image, s.tag, args))
	return errors.Wrap(err, output.String())
}

//go:embed resources/service/setup-service.sh
var setupService string

func (s *Service) setupService() error {
	_, output, err := s.code.bash(fmt.Sprintf(setupService, s.code.DB.Root, s.Name))
	return errors.Wrap(err, output.String())
}

func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}

	return s.Start()
}

func (s *Service) Stop() error {
	if !s.Running() {
		return nil
	}

	if _, _, err := s.code.docker("stop", s.Name); err != nil {
		return errors.Wrap(err, "failed to stop service")
	}

	if _, _, err := s.code.docker("rm", s.Name); err != nil {
		return errors.Wrap(err, "failed to remove service")
	}

	return nil
}

func (s *Service) Proxy(prefix string) http.Handler {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", s.Port))
	if err != nil {
		log.Fatal("Failed to create reverse proxy: ", err)
	}

	h := httputil.NewSingleHostReverseProxy(url)
	return http.StripPrefix(prefix, h)
}
