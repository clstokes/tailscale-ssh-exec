# tailscale-ssh-exec

[![status: experimental](https://img.shields.io/badge/status-experimental-blue)](https://tailscale.com/kb/1167/release-stages/#experimental)

A program to wrap for shell access via Tailscale SSH to restrict the commands that can be run based on the remote Tailscale user that is connecting.

## Usage

1. Build the `tailscale-ssh-exec` binary with `GOOS=linux go build -o tailscale-ssh-exec main.go`.
1. Install the `tailscale-ssh-exec` binary from the previous step and `tailscale-ssh-exec.sh` somewhere on your server - i.e. `/usr/local/bin/`.
1. Ensure both files are readable and executable by any user.

    ```shell
    chmod 755 /usr/local/bin/tailscale-ssh-exec /usr/local/bin/tailscale-ssh-exec.sh
    ```

1. Modify `/etc/passwd` on your server to run `tailscale-ssh-exec.sh` as the shell for users you need to control commands for.

    ```shell
    hg:x:1001:1001::/home/hg:/usr/local/bin/tailscale-ssh-exec.sh
    ```

1. Restrict port forwading, sftp, and pty via Tailscale environment variables. Typically you would do this by adding the following to `/etc/default/tailscaled` with:

    ```shell
    TS_SSH_DISABLE_SFTP=true
    TS_SSH_DISABLE_FORWARDING=true
    TS_SSH_DISABLE_PTY=true
    ```

1. Create a CSV that maps Tailscale login names to the specific command and arguments to run for each user.

    ```csv
    user1@example.com,/usr/bin/echo Hi user1
    user2@example.com,/usr/bin/hg-ssh /home/hg/repo
    ```

1. Modify `tailscale-ssh-exec.sh` to specify the CSV file from the previous step and customize any additional arguments to `tailscale-ssh-exec` - e.g. `-v` to enable verbose logging; helpful for troubleshooting during set up.

```shell
#!/bin/sh

/usr/local/bin/tailscale-ssh-exec \
    "$@" \
    -tailscale-ssh-exec-user-commands-file /tmp/example-user-to-commands.csv \
    -tailscale-ssh-exec-verbose
```
