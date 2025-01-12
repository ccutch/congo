package main

import (
	"cmp"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_chat"

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
		congo_auth.WithSetupView("signup.html", "/"),
		congo_auth.WithSigninView("user", "signin.html"),
		congo_auth.WithSignupCallback(signup))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "chatter.db", migrations)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "forest")),
		// congo.WithHostPrefix("/coder/proxy/8000"),
		congo.WithTemplates(congo_chat.Templates),
		congo.WithTemplates(templates),
		congo.WithController(auth.Controller()),
		congo.WithController("chatting", new(controllers.ChattingController)),
		congo.WithController("settings", new(controllers.SettingsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	app.Handle("/", app.Serve("homepage.html"))
	app.Handle("/{user}", auth.Protect(app.Serve("messages.html")))

	app.StartFromEnv()
}

func signup(auth *congo_auth.Controller, user *congo_auth.Identity) http.HandlerFunc {
	chatting := auth.Use("chatting").(*controllers.ChattingController)

	return func(w http.ResponseWriter, r *http.Request) {
		name := fmt.Sprintf("%s's Mailbox", user.Username)
		if _, err := chatting.Chat.NewMailboxWithID(user.ID, user.ID, name, 100); err != nil {
			auth.Render(w, r, "error-message", err)
			return
		}
		auth.Redirect(w, r, "/me")
	}
}
