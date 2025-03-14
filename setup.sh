echo Installing global packages...
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux certbot build-essential npm

echo Installing golang from source...
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz && \
    sudo rm -rf /usr/local/go && \
    sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz && \
rm go1.24.0.linux-amd64.tar.gz

echo Updating Bash environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin'      >> $HOME/.bashrc
source $HOME/.bashrc