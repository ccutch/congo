docker run -d \
		--name %[1]s \
		--net host \
		-e PORT=%[3]d \
		-p %[3]d:%[3]d \
    -v %[2]s/workspace/%[1]s/.config:/home/coder/.config \
		-v %[2]s/workspace/%[1]s/project:/home/coder/project \
		codercom/code-server:latest \
		--auth none