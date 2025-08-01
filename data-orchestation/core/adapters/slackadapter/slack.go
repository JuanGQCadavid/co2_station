package slackadapter

import (
	"log"

	"github.com/slack-go/slack"
)

type SlackNotification struct {
	api       *slack.Client
	channelId string
}

func NewSlackNotification(token, channelId string) *SlackNotification {
	return &SlackNotification{
		api:       slack.New(token),
		channelId: channelId,
	}
}

func (slaki *SlackNotification) Send(msg string, logLevel string) {
	_, _, err := slaki.api.PostMessage(
		slaki.channelId,
		slack.MsgOptionText(msg, false),
	)

	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	log.Println("message sent successfully")
}
