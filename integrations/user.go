package integrations

import (
	"errors"

	"gopkg.in/underarmour/dynago.v1"
	"gopkg.in/underarmour/dynago.v1/schema"
)

const (
	batchLimit = 25

	emailKey     = "email"
	firstNameKey = "first_name"
	lastNameKey  = "last_name"
	slackKey     = "slack"
	phoneKey     = "phone"
	awsKey       = "aws"
	githubKey    = "github"

	slackIndex  = "slack-index"
	awsIndex    = "aws-index"
	githubIndex = "github-index"
)

var (
	// EmailIndex is used for querying Dynamo for a user based on their email. This is also
	// the primary index for Dynamo.
	EmailIndex = Index{"", "email"}
	// FreeTierThroughput is set low within the free tier
	FreeTierThroughput = schema.ProvisionedThroughput{
		ReadCapacityUnits:  2,
		WriteCapacityUnits: 2,
	}
	// ErrUserDNE represents the case when a query executes properly but the user is
	// not found in the database.
	ErrUserDNE = errors.New("User not found")
)

// Index represents the information needed to query Dynamo on a Global Secondary index
// for a certain field.
type Index struct {
	Index string
	Field string
}

// User represents the data collected and served by who's who
type User struct {
	FirstName string `json:"first_name"`       // Slack
	LastName  string `json:"last_name"`        // Slack
	Email     string `json:"email"`            // Slack
	Slack     string `json:"slack"`            // Slack
	Phone     string `json:"phone"`            // Slack
	AWS       string `json:"aws"`              // first initial + last name
	Github    string `json:"github,omitempty"` // Github
}

// ToDynago converts a User object into a dynago.Document object.
func (u User) ToDynago() dynago.Document {
	d := dynago.Document{
		emailKey:     u.Email,
		slackKey:     u.Slack,
		firstNameKey: u.FirstName,
		lastNameKey:  u.LastName,
		phoneKey:     u.Phone,
		awsKey:       u.AWS,
	}

	// don't overwrite Github keys with empty strings
	if u.Github != "" {
		d[githubKey] = u.Github
	}
	return d
}

// UserFromDynago builds a user object from a dynago.Document.
func UserFromDynago(doc dynago.Document) User {
	return User{
		Email:     doc.GetString(emailKey),
		FirstName: doc.GetString(firstNameKey),
		LastName:  doc.GetString(lastNameKey),
		Slack:     doc.GetString(slackKey),
		Phone:     doc.GetString(phoneKey),
		AWS:       doc.GetString(awsKey),
		Github:    doc.GetString(githubKey),
	}
}

// UserMap is used to flesh out the User objects with data from additional services.
// The string key will correspond to the primary email of each Google Apps user.
type UserMap map[string]User

// InfoSource represents a data source for Who's Who.
type InfoSource interface {
	// Fill adds this data source's attributes of the user. It is expected that a user may
	// not have information in every InfoSource.
	Fill(UserMap) (UserMap, error)
}
