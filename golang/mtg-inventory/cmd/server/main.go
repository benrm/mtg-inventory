/*
Executable server runs an instance of the slack.Server.
*/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/benrm/mtg-inventory/golang/mtg-inventory/slack"
)

func main() {
	var failed bool
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must be set.\n")
		failed = true
	} else if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must have the prefix \"xapp-\".\n")
		failed = true
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must be set.\n")
		failed = true
	} else if !strings.HasPrefix(botToken, "xoxb-") {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must have the prefix \"xoxb-\".\n")
		failed = true
	}

	if failed {
		os.Exit(1)
	}

	server := slack.NewServer(nil, nil, appToken, botToken)

	err := server.Serve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with server: %s\n", err.Error())
		os.Exit(1)
	}
}
