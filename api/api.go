package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	"github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/slack"
	"github.com/gorilla/mux"
)

// m is a convenience type for using kayvee.
type m map[string]interface{}

// DynamoConn wraps the Dynamo client and it's helper functions for Users.
type DynamoConn struct {
	Dynamo integrations.Client
}

// aliasEndpoint generates a HTTP handler that will query for a user with the specified
// dynamo index and provided value.
func (d DynamoConn) listEndpoint(w http.ResponseWriter, r *http.Request) {
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
func (d DynamoConn) aliasEndpoint(idx integrations.Index, key string) http.HandlerFunc {
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

// HookUpRouter sets up the router
func (d DynamoConn) HookUpRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/list", d.listEndpoint)
	r.HandleFunc("/alias/aws/{username}", d.aliasEndpoint(cleveraws.Index, "username"))
	r.HandleFunc("/alias/slack/{handle}", d.aliasEndpoint(slack.Index, "handle"))
	r.HandleFunc("/alias/email/{email}", d.aliasEndpoint(integrations.EmailIndex, "email"))
	return r
}
