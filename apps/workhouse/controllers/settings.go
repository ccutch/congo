package controllers

import (
	"cmp"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type SettingsController struct {
	congo.BaseController
}

func (settings *SettingsController) Setup(app *congo.Application) {
	auth := app.Use("auth").(*AuthController)

	settings.BaseController.Setup(app)
	app.Handle("POST /_settings/name", auth.ProtectFunc(settings.updateName, "developer"))
	app.Handle("POST /_settings/description", auth.ProtectFunc(settings.updateDescription, "developer"))
	app.Handle("POST /_settings/theme", auth.ProtectFunc(settings.updateTheme, "developer"))
}

func (settings SettingsController) Handle(req *http.Request) congo.Controller {
	settings.Request = req
	return &settings
}

func (settings *SettingsController) set(id, val string) error {
	return settings.DB.Query(`

		INSERT INTO settings (id, value)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP

	`, id, val).Exec()
}

func (settings *SettingsController) get(id string) (val string) {
	settings.DB.Query(`
	
		SELECT value
		FROM settings WHERE id = ?
	
	`, id).Scan(&val)
	return val
}

func (settings *SettingsController) Has(id string) bool {
	return settings.get(id) != ""
}

func (settings *SettingsController) Name() string {
	return cmp.Or(settings.get("name"), "Workhouse")
}

func (settings *SettingsController) Description() string {
	return settings.get("description")
}

func (settings *SettingsController) Theme() string {
	auth := settings.Use("auth").(*AuthController)
	i, _ := auth.Authenticate(settings.Request, "developer", "user")
	return settings.get(i.ID + ":theme")
}

func (settings *SettingsController) updateName(w http.ResponseWriter, r *http.Request) {
	settings.set("name", r.FormValue("name"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings *SettingsController) updateDescription(w http.ResponseWriter, r *http.Request) {
	settings.set("description", r.FormValue("description"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings SettingsController) updateTheme(w http.ResponseWriter, r *http.Request) {
	auth := settings.Use("auth").(*AuthController)
	i, _ := auth.Authenticate(r, "developer", "user")
	settings.set(i.ID+":theme", r.FormValue("theme"))
	w.WriteHeader(http.StatusNoContent)
}
