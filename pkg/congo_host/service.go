package congo_host

import (
	"bytes"
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
	Target
	Host  *CongoHost
	Name  string
	Port  int
	Image string
	Tag   string
	args  []string
	envs  []string
	vols  []string
}

func (s *LocalHost) Service(name string, opts ...ServiceOpt) *Service {
	info := &Service{s, s.host, name, 0, "", "latest", []string{}, []string{}, []string{}}
	for _, opt := range opts {
		opt(info)
	}
	return info
}

func (s *RemoteHost) Service(name string, opts ...ServiceOpt) *Service {
	info := &Service{s, s.host, name, 0, "", "latest", []string{}, []string{}, []string{}}
	for _, opt := range opts {
		opt(info)
	}
	return info
}

type ServiceOpt func(*Service)

func WithPort(port int) ServiceOpt {
	return func(s *Service) { s.Port = port }
}

func WithImage(image string) ServiceOpt {
	return func(s *Service) { s.Image = image }
}

func WithTag(tag string) ServiceOpt {
	return func(s *Service) { s.Tag = tag }
}

func WithArgs(args ...string) ServiceOpt {
	return func(s *Service) { s.args = args }
}

func WithEnv(name, value string) ServiceOpt {
	env := fmt.Sprintf("%s=%s", name, value)
	return func(s *Service) { s.envs = append(s.envs, env) }
}

func WithVolume(volume string) ServiceOpt {
	return func(s *Service) { s.vols = append(s.vols, volume) }
}

func (s *Service) Running() bool {
	var stdout bytes.Buffer
	s.SetStdout(&stdout)
	err := s.Run("docker", "inspect", "-f", "{{.State.Status}}", s.Name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
}

//go:embed resources/service/start-service.sh
var startService string

func (s *Service) Start() error {
	if s.Running() {
		return nil
	}

	if s.Image == "" {
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
	return s.Run("bash", "-c", fmt.Sprintf(startService, s.Name, s.Port, envs, volumes, s.Image, s.Tag, args))
}

//go:embed resources/service/setup-service.sh
var setupService string

func (s *Service) setupService() error {
	log.Println("service", s)
	log.Println("host", s.Host)
	log.Println("db", s.Host.DB)
	return s.Run("bash", "-c", fmt.Sprintf(setupService, s.Host.DB.Root, s.Name))
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

	if err := s.Run("docker", "stop", s.Name); err != nil {
		return errors.Wrap(err, "failed to stop service")
	}

	if err := s.Run("docker", "rm", s.Name); err != nil {
		return errors.Wrap(err, "failed to remove service")
	}

	return nil
}

func (s *Service) Proxy(prefix string) http.Handler {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", s.Port))
	if err != nil {
		log.Fatal("Failed to create reverse proxy: ", err)
	}

	var h http.Handler = httputil.NewSingleHostReverseProxy(url)
	if prefix != "" {
		h = http.StripPrefix(prefix, h)
	}
	return h
}
