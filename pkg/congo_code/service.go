package congo_code

import (
	_ "embed"
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

//go:embed resources/start-service.sh
var startService string

type Service struct {
	code *CongoCode
	Name string
	Port int
}

type ServiceOpt func(*Service) error

func (code *CongoCode) Service(name string, opts ...ServiceOpt) *Service {
	s := Service{code, name, 0}
	for _, opt := range opts {
		if err := opt(&s); err != nil {
			log.Fatal("Failed to setup CongoCode: ", err)
		}
	}
	return &s
}

func WithPort(port int) ServiceOpt {
	return func(s *Service) error {
		s.Port = port
		return nil
	}
}

func (s *Service) Start() error {
	if s.Running() {
		return nil
	}
	_, output, err := s.code.bash(fmt.Sprintf(startService, s.Name, s.code.DB.Root, s.Port))
	if err != nil {
		log.Printf("Service %s output: %s", s.Name, output.String())
		return errors.Wrap(err, "failed to start service")
	}
	return nil
}

func (s *Service) Running() bool {
	stdout, _, err := s.code.docker("inspect", "-f", "{{.State.Status}}", s.Name)
	return err == nil && strings.TrimSpace(stdout.String()) == "running"
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

func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Service) Proxy() *httputil.ReverseProxy {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", s.Port))
	if err != nil {
		log.Fatal("Failed to create reverse proxy: ", err)
	}
	return httputil.NewSingleHostReverseProxy(url)
}
