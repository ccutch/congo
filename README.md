# Congo
Welcome to my Go based MVC framework. Feel free to fork this repo and get started with the following.


## Installing Tools
The easiest way to get started is by installing the tools found in this repo's `cmd` directory like so:

```bash
go install github.com/ccutch/congo/...

create-congo-app # To get help writing code

congo-hosting # For help deploying your app
```


## Running the Project
To get started running the project locally use the following command:

```bash
go run ./example
```

#### Models
You can find a full example of a model in `example/models/post.go`.

```go

//...

type Post struct {
	congo.Model
	Title   string
	Content string
}

func SearchPosts(db *congo.Database, query string) (posts []*Post, err error) {
	return posts, db.Query(`
	
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE title LIKE ?
	
	`, "%"+query+"%").All(func(scan congo.Scanner) (err error) {

		p := Post{Model: congo.Model{DB: db}}
		posts = append(posts, &p)
		return scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt)

	})
}

//...

```

#### Controller
You can find an example of a controller in `example/controllers/posts.go`.

```go

//...

type PostController struct{ congo.BaseController }

func (posts *PostController) Setup(app *congo.Application) {
	posts.Application = app
	app.HandleFunc("POST /blog", posts.handleCreate)
	app.HandleFunc("PUT /blog/{post}", posts.handleUpdate)
}

func (posts PostController) Handle(r *http.Request) congo.Controller {
	posts.Request = r
	return &posts
}

func (posts *PostController) SearchPosts() ([]*models.Post, error) {
	return models.SearchPosts(posts.DB, posts.PathValue("query"))
}

//...

```

#### Views
You can find an example of a view in `example/templates/blog-posts.html`.

```html

<!-- ... -->

{{range posts.SearchPosts}}
<a href="{{host}}/blog/{{.ID}}" class="card bg-base-300 shadow">
    <div class="card-body">
        <h2 class="card-title">{{.Title}}</h2>
    </div>
</a>
{{end}}

<!-- ... -->

```

#### Runner
To run the app setup the app like shown in ./examples/main.go:
```go

//...

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-data")

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(path, "app.db", migrations)),
		congo.WithController("posts", new(controllers.PostController)),
		congo.WithTemplates(templates))
)

func main() {
	//...

	http.Handle("GET /blog", app.Serve("blog-posts.html"))

	congo_boot.StartFromEnv(app)
}

```

## Deploying to Digital Ocean
You can use the command line tool found in `cmd/congo-hosting` to quickly get setup hosting you own congo instance. Run the following commands:

```bash
export DIGITAL_OCEAN_API_KEY=<your-digital-ocean-api-key>

congo-hosting launch

curl http://<your-ip-address>:8080
```


#### Forwarding Traffic
First take the IP address given to you when the server was launch and create an A record with your DNS provider. Then run the following:
```bash
congo-hosting gen-certs --domain www.example.com

curl https://www.example.com
```


#### Connect to Server
If you want to connect to the server directly via a secure connection use:
```bash
congo-hosting connect

tmux ls # <-- check for the running server process in tmux
```


#### Restarting Server
If you need to restart the server, and maybe want to update the binary use:
```bash
go build -o congo . # Building a new binary executable
congo-hosting restart --binary ./congo
```