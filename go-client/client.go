package whoswho

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
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
	FirstName string `json:"first_name,omitempty"` // FirstName
	LastName  string `json:"last_name,omitempty"`  // LastName
	Email     string `json:"email"`                // Email
	Slack     string `json:"slack,omitempty"`      // Slack
	Phone     string `json:"phone,omitempty"`      // Phone
	SlackID   string `json:"slack_id,omitempty"`   // Slack ID (not Slack alias)
	AWS       string `json:"aws,omitempty"`        // first initial + last name
	Github    string `json:"github,omitempty"`     // Github username
	Active    bool   `json:"active,omitempty"`     // Is user currently at Clever
	Team      string `json:"team,omitempty"`       // What team is the user on
	Pickabot  `json:"pickabot,omitempty"`
}

// Pickabot is config specific to https://github.com/clever/pickabot
type Pickabot struct {
	TeamOverrides []PickabotTeamOverride `json:"team_overrides,omitempty"`
	Flair         string                 `json:"flair"`
}

// PickabotTeamOverride describes a temporary override when picking a team member
type PickabotTeamOverride struct {
	Team    string `json:"team"`
	Include bool   `json:"include"`
	Until   int64  `json:"until"` // unix timestamp in seconds
}

// GetUserList makes a call to /list and returns all users.
func (c Client) GetUserList() ([]User, error) {
	resp, err := retryablehttp.Get(c.endpoint + "/list")
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
	userInfoReader := bytes.NewReader(userInfoJson)

	httpClient := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequest("PUT", c.endpoint+fmt.Sprintf("/alias/email/%s", email), userInfoReader)
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
	resp, err := retryablehttp.Get(c.endpoint + fmt.Sprintf("/alias/aws/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("aws alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserByGithub finds a user based on their Github username.
func (c Client) UserByGithub(username string) (User, error) {
	resp, err := retryablehttp.Get(c.endpoint + fmt.Sprintf("/alias/github/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("aws alias match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserBySlack finds a user based on their Slack username.
func (c Client) UserBySlack(username string) (User, error) {
	resp, err := retryablehttp.Get(c.endpoint + fmt.Sprintf("/alias/slack/%s", username))
	if err != nil {
		return User{}, fmt.Errorf("slack alias match call failed => {%s}", err)
	}
	return returnUser(resp)
}

// UserByEmail finds a user based on their email.
func (c Client) UserByEmail(email string) (User, error) {
	resp, err := retryablehttp.Get(c.endpoint + fmt.Sprintf("/alias/email/%s", email))
	if err != nil {
		return User{}, fmt.Errorf("email match call failed => {%s}", err)
	}

	return returnUser(resp)
}

// UserBySlackID finds a user by SlackID
func (c Client) UserBySlackID(slackID string) (User, error) {
	resp, err := retryablehttp.Get(c.endpoint + fmt.Sprintf("/alias/slack_id/%s", slackID))
	if err != nil {
		return User{}, fmt.Errorf("slack ID match call failed => {%s}", err)
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
