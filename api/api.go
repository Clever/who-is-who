package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	"github.com/Clever/who-is-who/integrations/cleverAWS"
	"github.com/Clever/who-is-who/integrations/github"
	"github.com/Clever/who-is-who/integrations/slack"
	"github.com/gorilla/mux"
)

var (
	welcomePageBuffer   bytes.Buffer
	welcomePageTemplate = `
<h1>Who's Who</h1>

<p>Endpoints</p>

{{ range . }}
  <h3>{{ .Endpoint }}</h3>
  <p>{{ .Description }}</p>
{{ end }}`
	routesImplemented = []struct {
		Endpoint, Description string
	}{
		{"/alias/email/:email", "Returns info for a user with an email of :email"},
		{"/alias/slack/:handle", "Returns info for a user with a Slack handle of :handle"},
		{"/alias/aws/:username", "Returns info for a user with an AWS username of :username"},
		{"/alias/github/:username", "Returns info for a user with a Github username of :username"},
		{"/list", "Returns info for all user"},
	}
)

func init() {
	// execute and store the welcome page template to display routes available to API consumers
	tmpl, err := template.New("welcome").Parse(welcomePageTemplate)
	if err != nil {
		log.Fatal(kayvee.FormatLog("who-is-who", kayvee.Error, "template parsing error", m{
			"message": err.Error(),
		}))
	}
	err = tmpl.Execute(&welcomePageBuffer, routesImplemented)
	if err != nil {
		log.Fatal(kayvee.FormatLog("who-is-who", kayvee.Error, "template building error", m{
			"message": err.Error(),
		}))
	}

}

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
	r.HandleFunc("/alias/github/{username}", d.aliasEndpoint(github.Index, "username"))
	r.HandleFunc("/health/check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // 200 status is autoset
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for k := range r.Header {
			fmt.Println(k)
		}
		fmt.Printf("%#v\n", r.Header)
		w.Write(welcomePageBuffer.Bytes())
	})
	return r
}
