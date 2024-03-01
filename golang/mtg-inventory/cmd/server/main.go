/*
Executable server runs an instance of the slack.Server.
*/
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	backend "github.com/benrm/mtg-inventory/golang/mtg-inventory/backends/sql"
	"github.com/benrm/mtg-inventory/golang/mtg-inventory/scryfall"
	"github.com/benrm/mtg-inventory/golang/mtg-inventory/slack"
	_ "github.com/go-sql-driver/mysql"
)

var (
	bulkDataFile = flag.String("bulk_data", "./all-cards.json", "The bulk data file containing all Scryfall data")
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

	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %s\n", err.Error())
		os.Exit(1)
	}

	flag.Parse()

	bulkData, err := os.Open(*bulkDataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening bulk data file: %s\n", err.Error())
		os.Exit(1)
	}

	jsonCache, err := scryfall.NewJSONCache(bulkData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading bulk data file: %s\n", err.Error())
		os.Exit(1)
	}

	server := slack.NewServer(backend.NewBackend(db), jsonCache, appToken, botToken)

	err = server.Serve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with server: %s\n", err.Error())
		os.Exit(1)
	}
}
