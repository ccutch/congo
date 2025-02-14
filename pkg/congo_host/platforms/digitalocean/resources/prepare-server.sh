# Mount Droplet Volume
mkfs.ext4 /dev/disk/by-id/scsi-0DO_Volume_%[1]s-data
mkdir /mnt/data
mount -o defaults,nofail,discard,noatime /dev/disk/by-id/scsi-0DO_Volume_%[1]s-data /mnt/data

# Installing dependencies
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux gcc sqlite3

# Installing golang from source
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
    sudo rm -rf /usr/local/go && \
    sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
rm go1.23.2.linux-amd64.tar.gz

# Updating Bash environment
sed -i '1i export PATH=$PATH:/usr/local/go/bin' $HOME/.bashrc
sed -i '1i export PATH=$PATH:$HOME/go/bin' $HOME/.bashrc
sed -i '1i export CONGO_SERVER_NAME=%[1]s' $HOME/.bashrc
sed -i '1i export CONGO_SERVER_SIZE=%[2]s' $HOME/.bashrc
sed -i '1i export CONGO_HOST_REGION=%[3]s' $HOME/.bashrc
source $HOME/.bashrc


# Allow Firewall for 80 (Certbot) and 443 (SSL)
ufw allow 22
ufw allow 80
ufw allow 8080
ufw allow 443
ufw reload