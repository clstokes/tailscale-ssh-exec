package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"tailscale.com/client/tailscale"
	"tailscale.com/tailcfg"
)

func main() {

	f, err := os.OpenFile("/tmp/cameron.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		log.Println(fmt.Sprintf("%s = %s", pair[0], pair[1]))
	}

	// get user from LocalApi
	user, err := getTailscaleUserFromSshEnv(context.Background()) // TODO: is context.Background() correct?
	if err != nil {
		log.Println(fmt.Sprintf("error getting tailscale user [%v]", err))
		os.Exit(1)
	}

	// execute command
	userCommand := "/usr/bin/hg-ssh /home/hg/repo"
	userCommandSplit := strings.SplitN(userCommand, " ", -1)
	log.Println(fmt.Sprintf("connection from [%s] - running [%s] with args [%v]", user.LoginName, userCommandSplit[0], userCommandSplit[1:]))
	out, err := execCmd(userCommandSplit[0], userCommandSplit[1:])
	if err != nil {
		log.Println(fmt.Sprintf("error [%v] from command [%v]", err, out))
		os.Exit(1)
	}

	log.Println(fmt.Sprintf("command output [%v]", out))
}

func execCmd(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)

	// var out bytes.Buffer
	// cmd.Stdout = &out
	// cmd.Stderr = &out

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return "", nil
}

func getTailscaleUserFromSshEnv(ctx context.Context) (*tailcfg.UserProfile, error) {
	// SSH_CLIENT = 100.110.18.145 60800 22
	sshClient := os.Getenv("SSH_CLIENT")
	sshClientValues := strings.SplitN(sshClient, " ", 3)
	ipPort := fmt.Sprintf("%s:%s", sshClientValues[0], sshClientValues[1])

	user, err := getTailscaleUser(ctx, ipPort)
	if err != nil {
		return nil, fmt.Errorf("error getting Tailscale user: %v", err)
	}
	return user, nil
}

func getTailscaleUser(ctx context.Context, ipPort string) (*tailcfg.UserProfile, error) {
	localClient := &tailscale.LocalClient{}
	whois, err := localClient.WhoIs(ctx, ipPort)

	if err != nil {
		return nil, fmt.Errorf("failed to identify remote host: %w", err)
	}
	if whois.Node.IsTagged() {
		return nil, fmt.Errorf("tagged nodes do not have a user identity")
	}
	if whois.UserProfile == nil || whois.UserProfile.LoginName == "" {
		return nil, fmt.Errorf("failed to identify remote user")
	}

	return whois.UserProfile, nil
}
