/*
Package slack integrates the Backend and Scryfall into a Slack app.
*/
package slack

import (
	"context"
	"fmt"
	"log"
	"os"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Server contains everything necessary for a Slack app integrated with the
// Backend and Scryfall
type Server struct {
	Backend  inventory.Backend
	Scryfall inventory.Scryfall
	API      *slack.Client
	Client   *socketmode.Client
}

// NewServer returns a new Server
func NewServer(
	backend inventory.Backend,
	scryfall inventory.Scryfall,
	appToken string,
	botToken string,
) *Server {
	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	server := &Server{
		Backend:  backend,
		Scryfall: scryfall,
		API:      api,
		Client:   client,
	}
	return server
}

// Serve consumes events until it fails and returns an error
func (s *Server) Serve() error {
	go func() {
		for event := range s.Client.Events {
			switch event.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeSlashCommand:
				cmd, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("SlashCommand not a SlashCommand")
					continue
				}

				_, err := s.Backend.AddUserIfNotExist(context.Background(), cmd.UserID)
				if err != nil {
					payload := map[string]interface{}{
						"blocks": []slack.Block{
							slack.NewSectionBlock(
								&slack.TextBlockObject{
									Type: slack.PlainTextType,
									Text: fmt.Sprintf("Sorry, failed to add user <@%s>", cmd.UserID),
								},
								nil,
								nil,
							),
						},
					}
					s.Client.Ack(*event.Request, payload)
					continue
				}

				switch cmd.Command {
				case "/ping":
					payload := map[string]interface{}{
						"blocks": []slack.Block{
							slack.NewSectionBlock(
								&slack.TextBlockObject{
									Type: slack.PlainTextType,
									Text: "pong",
								},
								nil,
								nil,
							),
						},
					}
					s.Client.Ack(*event.Request, payload)
				default:
					log.Printf("Unhandled slash command: %s", cmd.Command)
				}
			default:
				log.Printf("Unhandled event type received: %s\n", event.Type)
			}
		}
	}()

	return s.Client.Run()
}
