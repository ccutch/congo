if tmux has-session -t congo-server 2>/dev/null; then
    tmux kill-session -t congo-server
fi

cp /root/congo /root/congo.d
tmux new-session -d -s congo-server "source /root/.bashrc && export DATA_PATH=/mnt/data && export PORT=%d && /root/congo.d"