package main

import (
	"log"
	"os"
	"strings"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	aws "github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/slack"
)

var (
	slackToken     string
	awsKey         string
	awsSecret      string
	dynamoTable    string
	dynamoEndpoint string
	dynamoRegion   string
	domain         string
)

// m is a convenience type for using kayvee.
type m map[string]interface{}

// requiredEnv tries to find a value in the environment variables. If a value is not
// found the program will panaic.
func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal(kayvee.FormatLog("who-is-who", kayvee.Error, "missing env var", m{
			"var": key,
		}))
	}
	return value
}

func init() {
	domain = requiredEnv("DOMAIN")
	slackToken = requiredEnv("SLACK_TOKEN")
	awsKey = requiredEnv("AWS_ACCESS_KEY")
	awsSecret = requiredEnv("AWS_SECRET_KEY")
	dynamoTable = requiredEnv("DYNAMO_TABLE")
	dynamoEndpoint = requiredEnv("DYNAMO_ENDPOINT")
	dynamoRegion = requiredEnv("DYNAMO_REGION")
}

func main() {
	// get DB conn
	client, err := integrations.NewClient(dynamoTable, dynamoEndpoint, dynamoRegion, awsKey, awsSecret)
	if err != nil {
		log.Fatalf("Failed to connect => {%s}", err)
	}

	// get Slack info
	slackData := slack.UserMap{
		Domain:  domain,
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

	// try to save everything
	err = client.SaveUsers(userMap)
	if err != nil {
		log.Fatalf("Failed to save users in Dynamo, %s", err)
	}

	log.Printf("Saved %d users", len(userMap))
}
