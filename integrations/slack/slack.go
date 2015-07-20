package slack

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Clever/who-is-who/integrations"
)

const (
	// slackListUsersEndpoint is the API endpoint to query for a list of all users.
	slackListUsersEndpoint = "https://slack.com/api/users.list"
)

// UserMap contains all users given by Slack in an API call. The key to the map is
// the email address.
type UserMap map[string]Member

// Init calls the Slack API and fills the map with all users.
// It is an idempotent method.
func (sul UserMap) Init(token string) error {
	// short circuit for repeated Init() calls
	if len(sul) > 0 {
		return nil
	}

	// make API call for all users
	resp, err := http.Get(slackListUsersEndpoint + fmt.Sprintf("?token=%s", token))
	if err != nil {
		return fmt.Errorf("Failed to make API call to Slack => {%s}", err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to get users list from Slack => {%d status}", resp.StatusCode)
	}
	defer resp.Body.Close()

	// parse response
	var l UserList
	err = json.NewDecoder(resp.Body).Decode(&l)
	if err != nil {
		return fmt.Errorf("Failed to parse Slack's response => {%s}", err)
	} else if !l.Ok {
		return fmt.Errorf("Response with %d members marked as not OK", len(l.Members))
	}

	// fill map with all real users' info
	for _, u := range l.Members {
		if !u.IsBot && !u.Deleted {
			sul[u.Profile.Email] = u
		}
	}

	return nil
}

// Fill adds all information that Slack is intended to provide to the User objects.
// This is [Email, SlackHandle, Names and Phone].
func (sul UserMap) Fill(m integrations.UserMap) integrations.UserMap {
	for email, user := range m {
		member, exists := sul[email]
		if exists {
			user.Email = member.Profile.Email
			user.Slack = member.Name
			user.FirstName = member.Profile.FirstName
			user.LastName = member.Profile.LastName
			user.Phone = member.Profile.Phone

			m[email] = user
		}
	}
	return m
}

// UserList represents the info returned for the user.list endpoint
type UserList struct {
	Members []Member `json:"members"`
	Ok      bool     `json:"ok"`
}

// Member represents Slack's record of a user.
type Member struct {
	Color             string `json:"color"`
	Deleted           bool   `json:"deleted"`
	HasFiles          bool   `json:"has_files"`
	ID                string `json:"id"`
	IsAdmin           bool   `json:"is_admin"`
	IsBot             bool   `json:"is_bot"`
	IsOwner           bool   `json:"is_owner"`
	IsPrimaryOwner    bool   `json:"is_primary_owner"`
	IsRestricted      bool   `json:"is_restricted"`
	IsUltraRestricted bool   `json:"is_ultra_restricted"`
	Name              string `json:"name"`
	Profile           struct {
		Email              string `json:"email"`
		FirstName          string `json:"first_name"`
		Image192           string `json:"image_192"`
		Image24            string `json:"image_24"`
		Image32            string `json:"image_32"`
		Image48            string `json:"image_48"`
		Image72            string `json:"image_72"`
		ImageOriginal      string `json:"image_original"`
		LastName           string `json:"last_name"`
		Phone              string `json:"phone"`
		RealName           string `json:"real_name"`
		RealNameNormalized string `json:"real_name_normalized"`
		Skype              string `json:"skype"`
		Title              string `json:"title"`
	} `json:"profile"`
	RealName string      `json:"real_name"`
	Status   interface{} `json:"status"`
	Tz       string      `json:"tz"`
	TzLabel  string      `json:"tz_label"`
	TzOffset int         `json:"tz_offset"`
}
