package whoswho

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Clever/who-is-who/integrations"
)

// Client represents the API client. It holds the endpoint the server is expected to be
// located at.
type Client struct {
	endpoint string
}

// NewClient crafts a new API client.
func NewClient(endpoint string) Client {
	return Client{endpoint}
}

// GetUserList makes a call to /list and returns all users.
func (c Client) GetUserList() ([]integrations.User, error) {
	resp, err := http.Get(c.endpoint + "/list")
	if err != nil {
		return []integrations.User{}, fmt.Errorf("list users call failed => {%s}", err)
	}

	defer resp.Body.Close()
	var users []integrations.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return []integrations.User{}, fmt.Errorf("json unmarshaling failed => {%s}", err)
	}

	return users, nil
}

// UserByAWS finds a user based on their AWS username.
func (c Client) UserByAWS(username string) (integrations.User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/aws/%s", username))
	if err != nil {
		return integrations.User{}, fmt.Errorf("aws alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserBySlack finds a user based on their Slack username.
func (c Client) UserBySlack(username string) (integrations.User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/slack/%s", username))
	if err != nil {
		return integrations.User{}, fmt.Errorf("slack alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserByEmail finds a user based on their email.
func (c Client) UserByEmail(email string) (integrations.User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/email/%s", email))
	if err != nil {
		return integrations.User{}, fmt.Errorf("email match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// returnUser performs necessary unmarshaling and response parsing for all alias endpoints.
func returnUser(resp *http.Response) (integrations.User, error) {
	if resp.StatusCode != http.StatusOK {
		return integrations.User{}, fmt.Errorf("%d status code", resp.StatusCode)
	}

	var u integrations.User
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return integrations.User{}, fmt.Errorf("json unmarshaling failed => {%s}", err)
	}

	return u, nil
}
