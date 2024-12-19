package controllers

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type SettingsController struct{ congo.BaseController }

func (settings *SettingsController) Setup(app *congo.Application) {
	settings.BaseController.Setup(app)
	app.HandleFunc("POST /_settings/theme", settings.updateTheme)
	app.HandleFunc("POST /_settings/token", settings.updateToken)
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

		SELECT value FROM settings WHERE id = ?

	`, id).Scan(&val)
	return val
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
	settings.set("theme", r.FormValue("theme"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings SettingsController) updateToken(w http.ResponseWriter, r *http.Request) {
	settings.set("token", r.FormValue("token"))
	w.WriteHeader(http.StatusNoContent)
}