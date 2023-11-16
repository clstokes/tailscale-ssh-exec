package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"tailscale.com/client/tailscale"
	"tailscale.com/tailcfg"
)

var (
	cmd = flag.String("c", "", "forced command via ssh - e.g. SSH_ORIGINAL_COMMAND")

	// These flags are prefixed with -tailscale-ssh-exec... to not interfere with any other arguments that may get passed in.
	userCommandMappingFile = flag.String("tailscale-ssh-exec-user-commands-file", "", "file containing user-to-command mapping - e.g. tailscale-login-name,command and args")
	verbose                = flag.Bool("tailscale-ssh-exec-verbose", false, "enable verbose logging")
)

func main() {
	flag.Parse()
	if *userCommandMappingFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// get user from LocalApi
	user, err := getTailscaleUserFromConnection(context.Background())
	if err != nil {
		logPrintln("error getting tailscale user [%v]", err)
		os.Exit(1)
	}

	// get command from cmdFile
	userCommand, err := findUserCommand(*userCommandMappingFile, user.LoginName)
	if err != nil {
		logPrintln("unable to find command for user [%v]", err)
		os.Exit(1)
	}

	// execute command
	userCommandSplit := strings.SplitN(userCommand, " ", -1)
	logPrintln("connection from [%s], incoming command [%v], running [%s] with args [%v]", user.LoginName, *cmd, userCommandSplit[0], userCommandSplit[1:])
	out, err := execCmd(userCommandSplit[0], userCommandSplit[1:], *cmd)
	if err != nil {
		logPrintln("error [%v] from command [%v]", err, out)
		os.Exit(1)
	}
}

func getTailscaleUserFromConnection(ctx context.Context) (*tailcfg.UserProfile, error) {
	// SSH_CLIENT = 100.110.18.145 60800 22
	sshClient := os.Getenv("SSH_CLIENT")
	sshClientValues := strings.SplitN(sshClient, " ", 3)
	ipPort := fmt.Sprintf("%s:%s", sshClientValues[0], sshClientValues[1])

	user, err := getTailscaleUserProfile(ctx, ipPort)
	if err != nil {
		return nil, fmt.Errorf("error getting Tailscale user: %v", err)
	}
	return user, nil
}

func findUserCommand(fileName string, loginName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lineSplit := strings.SplitN(line, ",", 2)
		for i := range lineSplit {
			lineSplit[i] = strings.TrimSpace(lineSplit[i])
		}

		if lineSplit[0] == loginName {
			cmd := lineSplit[1]
			return cmd, nil
		}
	}

	return "", fmt.Errorf("no command found for [%s]", loginName)
}

func execCmd(command string, args []string, ogCmd string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Env = append(os.Environ(), fmt.Sprintf("SSH_ORIGINAL_COMMAND=%s", ogCmd))

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return "", nil
}

func getTailscaleUserProfile(ctx context.Context, ipPort string) (*tailcfg.UserProfile, error) {
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

func logPrintln(format string, a ...any) {
	if *verbose == false {
		return
	}
	log.Println(fmt.Sprintf(format, a...))
}
