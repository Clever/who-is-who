package github

import (
	"fmt"

	"github.com/Clever/who-is-who/integrations"
	githubAPI "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// UserList represents an array of Membership records for a Github Organization.
type UserList struct {
	Org     string
	Members []githubAPI.Membership
}

// Init make the necessary API calls to get all members of a Github Org.
func (l UserList) Init(token string) error {
	// short-circuit to prevent needless API calls.
	if len(l.Members) > 0 {
		return nil
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	gh := githubAPI.NewClient(tc)

	lo := githubAPI.ListMembersOptions{}
	for {
		members, resp, err := gh.Organizations.ListMembers(l.Org, &lo)
		if err != nil {
			return fmt.Errorf("Failed to form HTTP request for Github => {%s}", err)
		}
		for _, m := range members {
			_ = m
			// TODO: decide if there is a work around we want to do here.
		}

		if resp.NextPage == 0 {
			break
		} else {
			lo.Page = resp.NextPage
		}
	}

	return nil
}

// Fill is meant to add github information to the map of User infos.
func (l UserList) Fill(m integrations.UserMap) integrations.UserMap {
	return m
}
