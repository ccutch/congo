# Mount Droplet Volume
mkfs.ext4 /dev/disk/by-id/scsi-0DO_Volume_%[1]s
mkdir /mnt/data
mount -o defaults,nofail,discard,noatime /dev/disk/by-id/scsi-0DO_Volume_%[1]s /mnt/data