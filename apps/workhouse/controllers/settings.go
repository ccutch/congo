package controllers

import (
	"cmp"
	"net/http"

	"github.com/ccutch/congo/apps/workhouse/models"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
)

type SettingsController struct {
	congo.BaseController
}

func (settings *SettingsController) Setup(app *congo.Application) {
	settings.BaseController.Setup(app)

	auth := app.Use("auth").(*AuthController)
	http.Handle("POST /_settings/name", auth.ProtectFunc(settings.updateName, "developer"))
	http.Handle("POST /_settings/description", auth.ProtectFunc(settings.updateDescription, "developer"))
	http.Handle("POST /_settings/theme", auth.ProtectFunc(settings.updateTheme, "developer"))
	http.Handle("POST /_settings/token", auth.ProtectFunc(settings.updateToken, "developer"))
	http.Handle("POST /_settings/skip-payments", auth.ProtectFunc(settings.skipPayments, "developer"))
	http.Handle("POST /_settings/hosting", auth.ProtectFunc(settings.updateHosting, "developer"))
}

func (settings SettingsController) Handle(req *http.Request) congo.Controller {
	settings.Request = req
	return &settings
}

func (settings *SettingsController) set(id, val string) error {
	return models.SetSetting(settings.DB, id, val)
}

func (settings *SettingsController) get(id string) (val string) {
	val, _ = models.GetSetting(settings.DB, id)
	return val
}

func (settings *SettingsController) Has(id string) bool {
	return settings.get(id) != ""
}

func (settings *SettingsController) Name() string {
	return cmp.Or(settings.get("name"), "Workhouse")
}

func (settings *SettingsController) DomainRoot() string {
	return settings.get("DOMAIN_ROOT")
}

func (settings *SettingsController) Description() string {
	return settings.get("description")
}

func (settings *SettingsController) Theme() string {
	auth := settings.Use("auth").(*AuthController)
	i, _ := auth.Authenticate(settings.Request, "developer", "user")
	return settings.get(i.ID + ":theme")
}

func (settings *SettingsController) IsSetup() bool {
	return settings.Has("HOST_API_KEY") &&
		settings.Has("STRIPE_ACCESS_TOKEN") &&
		settings.Has("HOST_SIZE") &&
		settings.Has("HOST_REGION") &&
		settings.Has("STORAGE_SIZE")
}

func (settings *SettingsController) IsStripeSetup() bool {
	token := settings.get("STRIPE_ACCESS_TOKEN")
	return token != "" && token != "skipped"
}

func (settings *SettingsController) HostSize() string {
	return settings.get("HOST_SIZE")
}

func (settings *SettingsController) HostRegion() string {
	return settings.get("HOST_REGION")
}

func (settings *SettingsController) StorageSize() string {
	return settings.get("STORAGE_SIZE")
}

func (settings SettingsController) updateName(w http.ResponseWriter, r *http.Request) {
	settings.set("name", r.FormValue("name"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings SettingsController) updateDescription(w http.ResponseWriter, r *http.Request) {
	settings.set("description", r.FormValue("description"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings SettingsController) updateTheme(w http.ResponseWriter, r *http.Request) {
	auth := settings.Use("auth").(*AuthController)
	i, _ := auth.Authenticate(r, "developer", "user")
	settings.set(i.ID+":theme", r.FormValue("theme"))
	w.WriteHeader(http.StatusNoContent)
}

func (settings SettingsController) updateToken(w http.ResponseWriter, r *http.Request) {
	settings.set("HOST_API_KEY", r.FormValue("api_key"))
	content := settings.Use("content").(*ContentController)
	content.Host.WithAPI(digitalocean.NewClient(r.FormValue("api_key")))
	settings.Refresh(w, r)
}

func (settings SettingsController) skipPayments(w http.ResponseWriter, r *http.Request) {
	settings.set("STRIPE_ACCESS_TOKEN", "skipped")
	settings.Refresh(w, r)
}

func (settings SettingsController) updateHosting(w http.ResponseWriter, r *http.Request) {
	settings.set("DOMAIN_ROOT", cmp.Or(r.FormValue("domain"), settings.get("DOMAIN_ROOT")))
	settings.set("HOST_SIZE", cmp.Or(r.FormValue("size"), settings.get("HOST_SIZE")))
	settings.set("HOST_REGION", cmp.Or(r.FormValue("region"), settings.get("HOST_REGION")))
	settings.set("STORAGE_SIZE", cmp.Or(r.FormValue("storage"), settings.get("STORAGE_SIZE")))
	settings.Refresh(w, r)
}
