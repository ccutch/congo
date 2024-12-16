docker run -d
		--name workspace
		--privileged
		--network services
    -v %[1]s/.config:/home/coder/.config
		-v %[1]s/project:/home/coder/project
		-v /var/run/docker.sock:/var/run/docker.sock
		codercom/code-server:latest
		--auth none