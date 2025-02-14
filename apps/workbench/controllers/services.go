package controllers

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ccutch/congo/apps/workbench/models"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host"
)

func Services(host *congo_host.CongoHost) (string, *ServicesController) {
	return "services", &ServicesController{host: host}
}

type ServicesController struct {
	congo.BaseController
	host *congo_host.CongoHost
}

func (services *ServicesController) Setup(app *congo.Application) {
	services.BaseController.Setup(app)
	http.HandleFunc("POST /_services/launch", services.handleLaunch)
	http.HandleFunc("POST /_services/{id}/restart", services.handleRestart)
	http.HandleFunc("DELETE /_services/{id}", services.handleDelete)

	allServices, _ := models.ListServices(services.DB)
	for _, s := range allServices {
		log.Println("Server started:", os.Getenv("CONGO_SERVER_NAME"), s.ID)
		http.Handle(fmt.Sprintf("%s.%s.congo.gg/", s.ID, services.HostName()), services.dynamicProxy(s.ID))
	}
}

func (services ServicesController) Handle(req *http.Request) congo.Controller {
	services.Request = req
	return &services
}

func (services *ServicesController) HostName() string {
	return os.Getenv("CONGO_SERVER_NAME")
}

func (services *ServicesController) Services() ([]*models.Service, error) {
	return models.ListServices(services.DB)
}

func (services ServicesController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	portInt, err := strconv.Atoi(r.FormValue("port"))
	if err != nil {
		services.Render(w, r, "error-message", err)
		return
	}

	s, err := models.NewService(services.DB, r.FormValue("name"), cmp.Or(r.FormValue("path"), "/"), portInt)
	if err != nil {
		services.Render(w, r, "error-message", err)
		return
	}

	go func() {
		coding := services.Use("coding").(*CodingController)
		dir, err := os.MkdirTemp("", "workbench-service-*")
		if err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		host := services.host.Local()
		host.SetStdout(os.Stdout)

		service := host.Service(s.ID,
			congo_host.WithImage(s.ID),
			congo_host.WithEnv("PORT", strconv.Itoa(s.Port)),
			congo_host.WithPort(s.Port))

		log.Println("Server started:", os.Getenv("CONGO_SERVER_NAME"), s.ID)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered:", r)
				}
			}()

			http.Handle(fmt.Sprintf("%s.%s.congo.gg/", s.ID, services.HostName()), services.dynamicProxy(s.ID))
		}()

		if err = coding.Repo.Copy(dir); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		dir = filepath.Join(dir, s.Path)
		if err = host.Run("nixpacks", "build", dir, "--name", s.ID); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		if err := service.Start(); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		s.Status = "started"
		s.Save()
	}()

	services.Refresh(w, r)
}

func (services ServicesController) dynamicProxy(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := models.GetService(services.DB, id)
		if err != nil {
			services.Render(w, r, "error-message", err)
			return
		}

		host := services.host.Local()
		service := host.Service(s.ID,
			congo_host.WithImage(s.ID),
			congo_host.WithEnv("PORT", strconv.Itoa(s.Port)),
			congo_host.WithPort(s.Port))

		service.Proxy("/").ServeHTTP(w, r)
	}
}

func (services ServicesController) handleRestart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, err := models.GetService(services.DB, id)
	if err != nil {
		services.Render(w, r, "error-message", err)
		return
	}

	s.Status = "restarting"
	s.Save()

	go func() {
		coding := services.Use("coding").(*CodingController)
		dir, err := os.MkdirTemp("", "workbench-service-*")
		if err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		if err = coding.Repo.Copy(dir); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		host := services.host.Local()
		host.SetStdout(os.Stdout)
		dir = filepath.Join(dir, s.Path)
		if err = host.Run("nixpacks", "build", dir, "--name", s.ID); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		service := host.Service(s.ID,
			congo_host.WithImage(s.ID),
			congo_host.WithEnv("PORT", strconv.Itoa(s.Port)),
			congo_host.WithPort(s.Port))

		if err := service.Start(); err != nil {
			s.Error = err.Error()
			s.Save()
			return
		}

		s.Status = "started"
		s.Save()
	}()

	services.Refresh(w, r)
}

func (services ServicesController) handleDelete(w http.ResponseWriter, r *http.Request) {
	s, err := models.GetService(services.DB, r.PathValue("id"))
	if err != nil {
		services.Render(w, r, "error-message", err)
		return
	}

	go services.host.Local().Service(s.Name).Stop()

	if err := s.Delete(); err != nil {
		s.Error = err.Error()
		s.Save()
		return
	}

	services.Refresh(w, r)
}
