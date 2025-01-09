certbot certonly --standalone -d %[1]s %[2]s --non-interactive --agree-tos --email connormccutcheon95@gmail.com --expand
exit_code=$?

if [ $exit_code -eq 0 ]; then
    cp /etc/letsencrypt/live/%[1]s/fullchain.pem /root/fullchain.pem
    cp /etc/letsencrypt/live/%[1]s/privkey.pem /root/privkey.pem
else
    exit $exit_code
fi