package github

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	githubAPI "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	kv "gopkg.in/clever/kayvee-go.v2"
)

// m is a convenience type for using kayvee.
type m map[string]interface{}

var (
	emailRgx *regexp.Regexp
	// Index specifies the data for querying with the Global Secondary Index created for
	// queries on Github usernames.
	Index = integrations.Index{
		Field: "github",
		Index: "github-index",
	}
)

// UserList represents an array of Membership records for a Github Organization.
type UserList struct {
	Token  string
	Domain string
	Org    string
}

// Fill make the necessary API calls to get all members of a Github Org. Then we attempt to find
// emails for every developer in their public history.
func (l UserList) Fill(u integrations.UserMap) (integrations.UserMap, error) {
	emailRgx = regexp.MustCompile(fmt.Sprintf(`"email":"([\w\.]+@%s)"`, l.Domain))

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: l.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	gh := githubAPI.NewClient(tc)

	lo := githubAPI.ListMembersOptions{}
	for {
		members, resp, err := gh.Organizations.ListMembers(l.Org, &lo)
		if err != nil {
			return u, fmt.Errorf("Failed to form HTTP request for Github => {%s}", err)
		}
		for _, mbr := range members {
			if mbr.Login != nil && *mbr.Login != "" {
				email := findEmail(gh, *mbr.Login)
				if email == "" {
					continue
				}

				// add username to user if we find one with a matching email
				user, exists := u[email]
				if exists {
					user.Github = *mbr.Login
					u[email] = user
				} else {
					log.Println(kv.FormatLog("who-is-who", kv.Info, "mismatched email", m{
						"message": fmt.Sprintf("Found %s email but no user", l.Domain),
						"email":   email,
					}))
				}
			}
		}

		// cycle through all pages of org users
		if resp.NextPage == 0 {
			break
		} else {
			lo.Page = resp.NextPage
		}
	}

	return u, nil
}

func findEmail(c *githubAPI.Client, username string) string {
	events, resp, err := c.Activity.ListEventsPerformedByUser(username, true, nil)
	if err != nil {
		log.Println(kv.FormatLog("who-is-who", kayvee.Error, "Github API error", m{
			"msg": err.Error(),
		}))
		return ""
	} else if resp.StatusCode != http.StatusOK {
		log.Println(kv.FormatLog("who-is-who", kayvee.Error, "Github API error", m{
			"status code": resp.StatusCode,
		}))
		return ""
	}

	for _, e := range events {
		if e.RawPayload != nil {
			matches := emailRgx.FindAllStringSubmatch(string(*e.RawPayload), 1)
			if len(matches) == 1 && len(matches[0]) == 2 {
				return strings.ToLower(matches[0][1])
			}
		}
	}

	return ""
}
