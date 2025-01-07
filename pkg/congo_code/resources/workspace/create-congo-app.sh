/usr/local/go/bin/go install github.com/ccutch/congo/cmd/...@latest
create-congo-app --name %[1]s --dest $HOME/project --template %[2]s

cd $HOME/project
git add .
git commit -m "Initial Commit"
git push origin master