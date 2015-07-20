package main

import (
	"log"
	"os"
	"strings"

	"github.com/Clever/who-is-who/integrations"
	aws "github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/slack"
)

const (
	slackTokenEnvKey = "SLACK_TOKEN"
)

var (
	slackToken string
)

func init() {
	slackToken = os.Getenv(slackTokenEnvKey)
	if slackToken == "" {
		log.Fatalf("slack token required to start app, please set '%s'", slackTokenEnvKey)
	}
}

func main() {
	// get DB conn
	client, err := integrations.NewClient("http://localhost:8000", "", "")
	if err != nil {
		log.Fatalf("Failed to connect => {%s}", err)
	}

	// get Slack info
	slackData := slack.UserMap{
		Domain:  "clever.com",
		Members: make(map[string]slack.Member),
	}
	err = slackData.Init(slackToken)
	if err != nil {
		log.Fatalf("Failed to initialize Slack user list => {%s}", err)
	}

	// seed a map of User's with emails
	userMap := make(map[string]integrations.User)
	for _, m := range slackData.Members {
		userMap[strings.ToLower(m.Profile.Email)] = integrations.User{}
	}

	// declare all data sources to be used
	dataSources := []struct {
		Service integrations.InfoSource
		Token   string
		Name    string
	}{
		{slackData, slackToken, "slack"},
		{aws.AwsService{}, "", "aws"},
	}

	// add data from every data source to every User object
	for _, src := range dataSources {
		err := src.Service.Init(src.Token)
		if err != nil {
			log.Printf("Failed to get data from source, %s => {%s}", src.Name, err)
			continue
		}
		userMap = src.Service.Fill(userMap)
	}
}
