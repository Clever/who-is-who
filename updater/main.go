package main

import (
	"log"
	"os"
	"strings"

	"github.com/Clever/who-is-who/integrations"
	aws "github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/slack"
	"gopkg.in/clever/kayvee-go.v2"
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
	slackData := slack.NewUserMap(domain, slackToken)

	// seed a map of User's with emails
	emails, err := slackData.EmailList()
	if err != nil {
		log.Fatal(kayvee.FormatLog("who-is-who", kayvee.Error, "bad slack conn", m{
			"msg": err.Error(),
		}))
	}

	userMap := make(integrations.UserMap)
	for _, e := range emails {
		userMap[strings.ToLower(e)] = integrations.User{}
	}

	// declare all data sources to be used
	dataSources := []struct {
		Service integrations.InfoSource
		Name    string
	}{
		{slackData, "slack"},
		{aws.AwsService{}, "aws"},
	}

	// add data from every data source to every User object
	for _, src := range dataSources {
		var err error
		userMap, err = src.Service.Fill(userMap)
		if err != nil {
			log.Printf("Failed to get data from source, %s => {%s}", src.Name, err)
		}
	}

	// try to save everything
	err = client.SaveUsers(userMap)
	if err != nil {
		log.Fatalf("Failed to save users in Dynamo, %s", err)
	}

	log.Printf("Saved %d users", len(userMap))
}
