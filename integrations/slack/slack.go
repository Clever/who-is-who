package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Clever/who-is-who/integrations"
)

var (
	// Index specifies the data for querying with the Global secondary index created for
	// queries on slack usernames.
	Index = integrations.Index{
		Index: "slack",
		Field: "slack",
	}
)

// slackListUsersEndpoint creates a full URL for Slack's list.user endpoint.
func slackListUserEndpoint(tkn string) string {
	qry := make(url.Values)
	qry.Set("token", tkn)
	//  is the API endpoint to query for a list of all users.
	return (&url.URL{
		Scheme:   "https",
		Host:     "slack.com",
		Path:     "/api/user.list",
		RawQuery: qry.Encode(),
	}).String()
}

// UserMap contains all users given by Slack in an API call. The key to the map is
// the email address.
type UserMap struct {
	members map[string]member
	domain  string
	token   string
}

// NewUserMap creates a new UserMap for obtaining data from Slack.
func NewUserMap(dmn, tkn string) UserMap {
	return UserMap{
		domain:  dmn,
		token:   tkn,
		members: make(map[string]member),
	}
}

// userList represents the info returned for the user.list endpoint
type userList struct {
	Members []member `json:"members"`
	Ok      bool     `json:"ok"`
}

// member represents Slack's record of a user.
type member struct {
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

// gatherData calls the Slack API and fills the map with all users.
func (sul UserMap) gatherData() error {
	// make API call for all users
	resp, err := http.Get(slackListUserEndpoint(sul.token))
	if err != nil {
		return fmt.Errorf("Failed to make API call to Slack => {%s}", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to get users list from Slack => {%d status}", resp.StatusCode)
	}

	// parse response
	var l userList
	err = json.NewDecoder(resp.Body).Decode(&l)
	if err != nil {
		return fmt.Errorf("Failed to parse Slack's response => {%s}", err)
	} else if !l.Ok {
		return fmt.Errorf("Response with %d members marked as not OK", len(l.Members))
	}

	// fill map with all real users' info
	for _, u := range l.Members {
		if u.Profile.Email != "" && u.Name != "" && !u.IsBot && !u.Deleted && strings.Contains(u.Profile.Email, sul.domain) {
			sul.members[strings.ToLower(u.Profile.Email)] = u
		}
	}

	return nil
}

// EmailList returns a list of all emails owned by this slack org's users.
func (sul UserMap) EmailList() ([]string, error) {
	if len(sul.members) == 0 {
		if err := sul.gatherData(); err != nil {
			return []string{}, err
		}
	}

	emails := make([]string, len(sul.members))
	var i int
	for e := range sul.members {
		emails[i] = e
		i++
	}
	return emails, nil
}

// Fill adds all information that Slack is intended to provide to the User objects.
// This is [Email, SlackHandle, Names and Phone].
func (sul UserMap) Fill(uMap integrations.UserMap) (integrations.UserMap, error) {
	if len(sul.members) == 0 {
		if err := sul.gatherData(); err != nil {
			return uMap, err
		}
	}

	for email, user := range uMap {
		m, exists := sul.members[email]
		if exists {
			user.Email = strings.ToLower(m.Profile.Email)
			user.Slack = m.Name
			user.FirstName = m.Profile.FirstName
			user.LastName = m.Profile.LastName
			user.Phone = m.Profile.Phone

			uMap[email] = user
		}
	}
	return uMap, nil
}
