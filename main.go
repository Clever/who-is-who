package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Clever/who-is-who/api"
	"github.com/Clever/who-is-who/integrations"
	"gopkg.in/clever/kayvee-go.v2"
)

var (
	port           string
	awsKey         string
	awsSecret      string
	dynamoTable    string
	dynamoRegion   string
	dynamoEndpoint string
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

func setupEnvVars() {
	port = requiredEnv("PORT")
	awsKey = requiredEnv("AWS_ACCESS_KEY_ID")
	awsSecret = requiredEnv("AWS_SECRET_ACCESS_KEY")
	dynamoTable = requiredEnv("DYNAMO_TABLE")
	dynamoRegion = requiredEnv("DYNAMO_REGION")
	dynamoEndpoint = requiredEnv("DYNAMO_ENDPOINT")
}

func main() {
	setupEnvVars()

	// setup dynamodb connection
	c, err := integrations.NewClient(dynamoTable, dynamoEndpoint, dynamoRegion, awsKey, awsSecret)
	if err != nil {
		log.Fatal(kayvee.FormatLog("who-is-who", kayvee.Error, "dynamo connection", m{
			"message": err.Error(),
		}))
	}
	d := api.DynamoConn{
		Dynamo: c,
	}

	// setup HTTP server
	log.Println(kayvee.FormatLog("who-is-who", kayvee.Info, "server startup", m{
		"message": fmt.Sprintf("Listening on %s", port),
	}))
	http.ListenAndServe(port, d.HookUpRouter())
}
