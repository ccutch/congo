# Install Golang
sudo apt-get update && \
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux && \
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
sudo rm -rf /usr/local/go && \
sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.bashrc

source $HOME/.bashrc