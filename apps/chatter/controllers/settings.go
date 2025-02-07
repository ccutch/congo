package controllers

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

type SettingsController struct {
	congo.BaseController
}

func (settings *SettingsController) Setup(app *congo.Application) {
	settings.BaseController.Setup(app)
	auth := app.Use("auth").(*congo_auth.AuthController)

	http.HandleFunc("POST /_settings/theme", auth.ProtectFunc(settings.updateTheme, "user"))
}

func (settings SettingsController) Handle(req *http.Request) congo.Controller {
	settings.Request = req
	return &settings
}

func (settings *SettingsController) Has(id string) bool {
	return settings.Get(id) != ""
}

func (settings *SettingsController) Get(id string) (val string) {
	settings.DB.Query(`
	
		SELECT value
		FROM settings WHERE id = ?
	
	`, id).Scan(&val)
	return val
}

func (settings *SettingsController) MyTheme() string {
	auth := settings.Use("auth").(*congo_auth.AuthController)
	i, _ := auth.Authenticate(settings.Request, "user")
	if i == nil {
		return ""
	}
	return settings.Get(i.ID + "-theme")
}

func (settings *SettingsController) set(id, val string) error {
	return settings.DB.Query(`

		INSERT INTO settings (id, value)
		VALUES ($1, $2)
		ON CONFLICT(id) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP

	`, id, val).Exec()
}

func (settings SettingsController) updateTheme(w http.ResponseWriter, r *http.Request) {
	auth := settings.Use("auth").(*congo_auth.AuthController)
	i, _ := auth.Authenticate(r, "user")
	settings.set(i.ID+"-theme", r.FormValue("theme"))
	w.WriteHeader(http.StatusNoContent)
}
