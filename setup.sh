# Installing dependencies
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux certbot build-essential

# Installing golang from source
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
    sudo rm -rf /usr/local/go && \
    sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
rm go1.23.2.linux-amd64.tar.gz

# Updating Bash environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin'      >> $HOME/.bashrc
source $HOME/.bashrc