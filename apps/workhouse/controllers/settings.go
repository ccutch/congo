package controllers

import (
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

type SettingsController struct {
	congo.BaseController
}

func (settings *SettingsController) Setup(app *congo.Application) {
	settings.BaseController.Setup(app)
	auth := app.Use("auth").(*congo_auth.Controller)

	app.HandleFunc("POST /_settings/theme", auth.ProtectFunc(settings.updateTheme))

	if settings.Get("token") == "" {
		settings.set("token", os.Getenv("DIGITAL_OCEAN_API_KEY"))
	}
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

func (settings *SettingsController) set(id, val string) error {
	return settings.DB.Query(`

		INSERT INTO settings (id, value)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP

	`, id, val).Exec()
}

func (settings SettingsController) updateTheme(w http.ResponseWriter, r *http.Request) {
	settings.set("theme", r.FormValue("theme"))
	w.WriteHeader(http.StatusNoContent)
}
