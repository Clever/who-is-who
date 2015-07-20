package integrations

// User represents the data collected and served by who's who
type User struct {
	FirstName string `json:"first_name"` // Slack
	LastName  string `json:"last_name"`  // Slack
	Email     string `json:"email"`      // Slack
	Slack     string `json:"slack"`      // Slack
	Phone     string `json:"phone"`      // Slack
	Github    string `json:"github"`     // Github
	AWS       string `json:"aws"`        // first initial + last name
}

// UserMap is used to flesh out the User objects with data from additional services.
// The string key will correspond to the primary email of each Google Apps user.
type UserMap map[string]User

// InfoSource represents a data source for Who's Who.
type InfoSource interface {
	// Init makes the necssary API calls to create a corpus of this data source's information.
	Init(token string) error
	// Fill adds this data source's attributes of the user. It is expected that a user may
	// not have information in every InfoSource.
	Fill(UserMap) UserMap
}
