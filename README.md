# Congo
Welcome to my Go based MVC framework. Feel free to fork this repo and get started with the following.


## Installing Tools
The easiest way to get started is by installing the tools found in this repo's `cmd` directory like so:
```
go install github.com/ccutch/congo/...

# This will install three binaries to `$HOME/go/bin`
congo # To run the example congo server

create-congo-app # To get writing code

congo-hosting # For deploying your app
```


## Running the Project
To get started running the project locally use the following command:
```
go run .
```

#### Models
You can find an example of a model in `models/post.go`.

#### Views
You can find an example of a view in `templates/blog-posts.html`.

#### Controller
You can find an example of a controller in `controllers/posts.go`.


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