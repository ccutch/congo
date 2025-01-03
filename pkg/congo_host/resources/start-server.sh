if tmux has-session -t congo-server 2>/dev/null; then
    tmux kill-session -t congo-server
fi

cp /root/congo /root/congo.d
tmux new-session -d -s congo-server "DATA_PATH=/mnt/data PORT=%d /root/congo.d"
