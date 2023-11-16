#!/bin/sh

/usr/local/bin/tailscale-ssh-exec \
    "$@" \
    -tailscale-ssh-exec-user-commands-file /etc/tailscale-ssh-exec/example-user-to-commands.csv \
    -tailscale-ssh-exec-verbose
