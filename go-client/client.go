package whoswho

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// User represents the data collected and served by who's who
type User struct {
	FirstName string `json:"first_name"`       // Slack
	LastName  string `json:"last_name"`        // Slack
	Email     string `json:"email"`            // Slack
	Slack     string `json:"slack"`            // Slack
	Phone     string `json:"phone"`            // Slack
	SlackID   string `json:"slack_id"`         // Slack
	AWS       string `json:"aws"`              // first initial + last name
	Github    string `json:"github,omitempty"` // Github
	Active    bool   `json:"active"`           // Is user currently at Clever
}

// GetUserList makes a call to /list and returns all users.
func (c Client) GetUserList() ([]User, error) {
	resp, err := http.Get(c.endpoint + "/list")
	if err != nil {
		return []User{}, fmt.Errorf("list users call failed => {%s}", err)
	}

	defer resp.Body.Close()
	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return []User{}, fmt.Errorf("json unmarshaling failed => {%s}", err)
	}

	return users, nil
}

// UpsertUser makes a PUT request to /alias/email/<email>, creates or updates the user, and returns the user
func (c Client) UpsertUser(author string, userInfo User) (User, error) {
	userInfoJson, err := json.Marshal(userInfo)
	email := userInfo.Email
	if err != nil {
		return User{}, fmt.Errorf("json marshaling failed => {%s}", err)
	}
	userInfoBuffer := bytes.NewBuffer(userInfoJson)

	httpClient := &http.Client{}
	req, err := http.NewRequest("PUT", c.endpoint+fmt.Sprintf("/alias/email/%s", email), userInfoBuffer)
	req.Header.Add("X-WIW-Author", author)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("add user call failed with email %s => {%s}", email, err)
	}

	return returnUser(resp)

}

// UserByAWS finds a user based on their AWS username.
func (c Client) UserByAWS(username string) (User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/aws/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("aws alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserByGithub finds a user based on their Github username.
func (c Client) UserByGithub(username string) (User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/github/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("aws alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserBySlack finds a user based on their Slack username.
func (c Client) UserBySlack(username string) (User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/slack/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("slack alias match call failed => {%s}", err)
	}
	return returnUser(resp)
}

// UserByEmail finds a user based on their email.
func (c Client) UserByEmail(email string) (User, error) {
	resp, err := http.Get(c.endpoint + fmt.Sprintf("/alias/email/%s", email))
	if err != nil {
		return User{}, fmt.Errorf("email match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// returnUser performs necessary unmarshaling and response parsing for all alias endpoints.
func returnUser(resp *http.Response) (User, error) {
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("%d status code", resp.StatusCode)
	}

	var u User
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return User{}, fmt.Errorf("json unmarshaling failed => {%s}", err)
	}

	return u, nil
}
