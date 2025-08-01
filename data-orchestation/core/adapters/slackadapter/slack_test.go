package slackadapter

import "testing"

func TestCall(t *testing.T) {

	theNoty := NewSlackNotification("--", "--")

	theNoty.Send("Hello there", "DEBUG")
	//
}
