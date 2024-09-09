#!/bin/sh

set -e

IP=$1

if [ -z "$IP" ]; then
  echo "Usage: $0 [unit ip address]" >&2
  exit 1
fi

GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -o r36s-demo-app .
du -h r36s-demo-app
sshpass -p root ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@$IP "killall r36s-demo-app || true"
sshpass -p root scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null r36s-demo-app root@$IP:/tmp/
sshpass -p root ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@$IP "sh -c 'cd /tmp; ./r36s-demo-app'"
