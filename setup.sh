# Installing dependencies
apt-get update
apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux certbot

# Installing golang from source
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
rm go1.23.2.linux-amd64.tar.gz

# Updating Bash environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin'      >> $HOME/.bashrc
source $HOME/.bashrc