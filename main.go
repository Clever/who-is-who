package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Clever/who-is-who/api"
	"github.com/Clever/who-is-who/integrations"
	kv "gopkg.in/Clever/kayvee-go.v2"
)

var (
	awsKey         string
	awsSecret      string
	dynamoTable    string
	dynamoRegion   string
	dynamoEndpoint string
	port           string
)

// m is a convenience type for using kv.
type m map[string]interface{}

func init() {
	flag.StringVar(&port, "port", ":80", "specify the HTTP port to listen on")
	flag.Parse()
}

// requiredEnv tries to find a value in the environment variables. If a value is not
// found the program will panaic.
func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal(kv.FormatLog("who-is-who", kv.Error, "missing env var", m{
			"var": key,
		}))
	}
	return value
}

func setupEnvVars() {
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
		log.Fatal(kv.FormatLog("who-is-who", kv.Error, "dynamo connection", m{
			"message": err.Error(),
		}))
	}
	d := api.DynamoConn{
		Dynamo: c,
	}

	// setup HTTP server
	log.Println(kv.FormatLog("who-is-who", kv.Info, "server startup", m{
		"message": fmt.Sprintf("Listening on %s", port),
	}))
	err = http.ListenAndServe(port, d.HookUpRouter())
	if err != nil {
		log.Fatal(kv.FormatLog("who-is-who", kv.Error, "server startup failure", m{
			"msg": err.Error(),
		}))
	}
}
