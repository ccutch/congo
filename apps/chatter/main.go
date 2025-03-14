package main

import (
	"bytes"
	"cmp"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/yuin/goldmark"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ccutch/congo/apps/chatter/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:public
	public embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("chatter"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("user", "signup.html"),
		congo_auth.WithSigninDest("/me"),
		congo_auth.WithSignupCallback(signup))

	app = congo.NewApplication(templates,
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_PATH"), "forest")),
		congo.WithFunc("markdown", markdown),
		congo.WithDatabase(congo.SetupDatabase(data, "chatter.db", migrations)),
		congo.WithController(auth.Controller()),
		congo.WithController("chatting", new(controllers.ChattingController)),
		congo.WithController("settings", new(controllers.SettingsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	http.Handle("/{$}", app.Serve("homepage.html"))
	http.Handle("/signin", app.Serve("signin.html"))
	http.Handle("/signup", app.Serve("signup.html"))
	http.Handle("/{user}", auth.Protect(app.Serve("messages.html"), "user"))
	http.Handle("/invite", auth.Protect(app.Serve("url-copied-toast"), "user"))
	http.Handle("/public/", http.FileServerFS(public))

	app.StartFromEnv()
}

func signup(auth *congo_auth.AuthController, user *congo_auth.Identity) http.HandlerFunc {
	chatting := auth.Use("chatting").(*controllers.ChattingController)
	return func(w http.ResponseWriter, r *http.Request) {
		name := fmt.Sprintf("%s's Mailbox", cases.Title(language.English).String(user.Name))
		if _, err := chatting.Chat.NewMailboxWithID(user.ID, user.ID, name); err != nil {
			auth.Render(w, r, "error-message", err)
			return
		}
		auth.Redirect(w, r, "/")
	}
}

func markdown(s string) template.HTML {
	var buf bytes.Buffer
	goldmark.Convert([]byte(s), &buf)
	return template.HTML(cmp.Or(buf.String(), s))
}
