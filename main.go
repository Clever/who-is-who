package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	"github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/slack"
	"github.com/gorilla/mux"
)

var (
	port           string
	awsKey         string
	awsSecret      string
	dynamoTable    string
	dynamoEndpoint string
	dynamoRegion   string
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
	awsKey = requiredEnv("AWS_ACCESS_KEY")
	awsSecret = requiredEnv("AWS_SECRET_KEY")
	dynamoTable = requiredEnv("DYNAMO_TABLE")
	dynamoEndpoint = requiredEnv("DYNAMO_ENDPOINT")
	dynamoRegion = requiredEnv("DYNAMO_REGION")
}

// dynamoConn wraps the Dynamo client and it's helper functions for Users.
type dynamoConn struct {
	Dynamo integrations.Client
}

// aliasEndpoint generates a HTTP handler that will query for a user with the specified
// dynamo index and provided value.
func (d dynamoConn) listEndpoint(w http.ResponseWriter, r *http.Request) {
	// query Dynamo for all users
	users, err := d.Dynamo.GetUserList()
	if err != nil {
		log.Println(kayvee.FormatLog("who-is-who", kayvee.Error, "getUserList error", m{
			"message": err.Error(),
		}))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write the user to the connection
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Println(kayvee.FormatLog("who-is-who", kayvee.Error, "json encoding error", m{
			"message": err.Error(),
		}))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// aliasEndpoint generates a HTTP handler that will query for a user with the specified
// dynamo index and provided value.
func (d dynamoConn) aliasEndpoint(idx integrations.Index, key string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the username/alias from the URL parameter
		username, valid := mux.Vars(r)[key]
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// query Dynamo for the user with the specified details.
		user, err := d.Dynamo.GetUser(idx, username)
		if err == integrations.ErrUserDNE {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			log.Println(kayvee.FormatLog("who-is-who", kayvee.Error, "getUser error", m{
				"message": err.Error(),
			}))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// write the user to the connection
		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			log.Println(kayvee.FormatLog("who-is-who", kayvee.Error, "json encoding error", m{
				"message": err.Error(),
			}))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// hookUpRouter sets up the router
func hookUpRouter(d dynamoConn) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/list", d.listEndpoint)
	r.HandleFunc("/alias/aws/{username}", d.aliasEndpoint(cleveraws.Index, "username"))
	r.HandleFunc("/alias/slack/{username}", d.aliasEndpoint(slack.Index, "username"))
	r.HandleFunc("/alias/email/{email}", d.aliasEndpoint(integrations.EmailIndex, "email"))
	return r
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
	d := dynamoConn{
		Dynamo: c,
	}

	// setup HTTP server
	log.Println(kayvee.FormatLog("who-is-who", kayvee.Info, "server startup", m{
		"message": fmt.Sprintf("Listening on %s", port),
	}))
	http.ListenAndServe(port, hookUpRouter(d))
}
