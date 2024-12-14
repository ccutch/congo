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

//...

```

#### Controller
You can find an example of a controller in `example/controllers/posts.go`.

```go

//...

type PostController struct{ congo.BaseController }

func (ctrl *PostController) Setup(app *congo.Application) {
	ctrl.Application = app
	app.HandleFunc("POST /blog", ctrl.handleCreate)
	app.HandleFunc("PUT /blog/{post}", ctrl.handleUpdate)
}

func (ctrl PostController) Handle(r *http.Request) congo.Controller {
	ctrl.Request = r
	return &ctrl
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

## Deploying to Digital Ocean
You can use the command line tool found in `cmd/congo-hosting` to quickly get setup hosting you own congo instance. Run the following commands:

```bash
export DIGITAL_OCEAN_API_KEY=<your-digital-ocean-api-key>

go run ./cmd/congo-hosting launch

curl http://<your-ip-address>:8080
```


#### Forwarding Traffic
First take the IP address given to you when the server was launch and create an A record with your DNS provider. Then run the following:
```bash
go run ./cmd/congo-hosting gen-certs --domain www.example.com

curl https://www.example.com
```


#### Connect to Server
If you want to connect to the server directly via a secure connection use:
```bash
go run ./cmd/congo-hosting connect

tmux ls # <-- check for the running server process in tmux
```


#### Restarting Server
If you need to restart the server, and maybe want to update the binary use:
```bash
go build -o congo . # Building a new binary executable
go run ./cmd/congo-hosting restart --binary ./congo
```