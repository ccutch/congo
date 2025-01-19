package main

import (
	"cmp"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ccutch/congo/apps/chatter/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("chatter"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("user", "signup.html"),
		congo_auth.WithSigninDest("/me"),
		congo_auth.WithSignupCallback(signup))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "chatter.db", migrations)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_PATH"), "forest")),
		congo.WithTemplates(templates),
		congo.WithController(auth.Controller()),
		congo.WithController("chatting", new(controllers.ChattingController)),
		congo.WithController("settings", new(controllers.SettingsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	app.Handle("/{$}", app.Serve("homepage.html"))
	app.Handle("/signin", app.Serve("signin.html"))
	app.Handle("/signup", app.Serve("signup.html"))
	app.Handle("/{user}", auth.Protect(app.Serve("messages.html")))

	app.StartFromEnv()
}

func signup(auth *congo_auth.AuthController, user *congo_auth.Identity) http.HandlerFunc {
	chatting := auth.Use("chatting").(*controllers.ChattingController)

	return func(w http.ResponseWriter, r *http.Request) {
		name := fmt.Sprintf("%s's Mailbox", cases.Title(language.English).String(user.Name))
		if _, err := chatting.Chat.NewMailboxWithID(user.ID, user.ID, name, 100); err != nil {
			auth.Render(w, r, "error-message", err)
			return
		}
		auth.Redirect(w, r, "/me")
	}
}
