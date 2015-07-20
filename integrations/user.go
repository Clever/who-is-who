package integrations

import (
	"fmt"
	"log"

	"github.com/underarmour/dynago"
	"github.com/underarmour/dynago/schema"
)

const (
	userTable  = "users"
	batchLimit = 25

	emailKey     = "email"
	firstNameKey = "first_name"
	lastNameKey  = "last_name"
	slackKey     = "slack"
	phoneKey     = "phone"
	awsKey       = "aws"
)

var (
	// EmailIndex is used for querying Dynamo for a user based on their email. This is also
	// the primary index for Dynamo.
	EmailIndex = Index{"", "email"}
	// FreeTierThroughput is the maximum throughput that we can use for Dynamo without
	// entering the paid tier.
	FreeTierThroughput = schema.ProvisionedThroughput{
		ReadCapacityUnits:  25,
		WriteCapacityUnits: 25,
	}
)

// Index represents the information needed to query Dynamo on a Global Secondary index
// for a certain field.
type Index struct {
	Index string
	Field string
}

// User represents the data collected and served by who's who
type User struct {
	FirstName string `json:"first_name"` // Slack
	LastName  string `json:"last_name"`  // Slack
	Email     string `json:"email"`      // Slack
	Slack     string `json:"slack"`      // Slack
	Phone     string `json:"phone"`      // Slack
	AWS       string `json:"aws"`        // first initial + last name
}

// ToDynago converts a User object into a dynago.Document object.
func (u User) ToDynago() dynago.Document {
	return dynago.Document{
		emailKey:     u.Email,
		slackKey:     u.Slack,
		firstNameKey: u.FirstName,
		lastNameKey:  u.LastName,
		phoneKey:     u.Phone,
		awsKey:       u.AWS,
	}
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
	}
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

// SaveUsers performs batch writes to add all users to Dynamo.
func (c Client) SaveUsers(l UserMap) error {
	dynagoObjects := make([]dynago.Document, len(l))
	var i int
	for _, u := range l {
		dynagoObjects[i] = u.ToDynago()
		i++
	}

	// do a batch write to Dynamo for every 25 users
	for i := 0; i < len(dynagoObjects)/batchLimit+1; i++ {
		// properly figure out the indexes on the array
		firstIndex := i * batchLimit
		lastIndex := (i + 1) * batchLimit
		if firstIndex >= len(dynagoObjects) {
			break
		} else if lastIndex > len(dynagoObjects) {
			lastIndex = len(dynagoObjects)
		}

		// perform write and check for unprocessed items
		res, err := c.Dynamo.BatchWrite().Put(userTable, dynagoObjects[firstIndex:lastIndex]...).Execute()
		if err != nil {
			return fmt.Errorf("Error while executing batch write: %s", err)
		} else if failedPuts := res.UnprocessedItems.GetPuts(userTable); len(failedPuts) > 0 {
			for _, fp := range failedPuts {
				log.Printf("Failed to store: {%#v}", fp)
			}
		}
	}

	return nil
}

// GetUser crafts a query for a single user based on the specified index and user information.
func (c Client) GetUser(idx Index, value string) (User, error) {
	res, err := c.Dynamo.
		Query(userTable).
		IndexName(idx.Index).
		KeyConditionExpression(fmt.Sprintf("%s = :username", idx.Field), dynago.Param{Key: ":username", Value: value}).
		Execute()

	if err != nil {
		return User{}, fmt.Errorf("Failed to make query with '%s'=='%s' due to: %s", idx.Field, value, err)
	} else if res.Count == 0 {
		return User{}, fmt.Errorf("Failed to find user with '%s'=='%s'", idx.Field, value)
	}

	return UserFromDynago(res.Items[0]), nil
}

// GetUserList returns all users.
func (c Client) GetUserList() ([]User, error) {
	res, err := c.Dynamo.Scan(userTable).Execute()
	if err != nil {
		return []User{}, fmt.Errorf("Failed to scan table, %s", err)
	}

	users := make([]User, len(res.Items))
	for i, d := range res.Items {
		users[i] = UserFromDynago(d)
	}

	return users, nil
}

// NewClient creates a conection to DynamoDB, then creates the
func NewClient(endpoint, region, accessKey, secretKey string) (Client, error) {
	executor := dynago.NewAwsExecutor(endpoint, region, accessKey, secretKey)
	client := dynago.NewClient(executor)

	// DescribeTable 400's if table DNE
	_, err := client.DescribeTable(userTable)
	if err != nil {
		_, err := client.CreateTable(&schema.CreateRequest{
			TableName: userTable,
			AttributeDefinitions: []schema.AttributeDefinition{
				{emailKey, schema.String},
				{slackKey, schema.String},
				{awsKey, schema.String},
			},
			KeySchema: []schema.KeySchema{
				{emailKey, schema.HashKey},
				{slackKey, schema.RangeKey},
			},
			ProvisionedThroughput: FreeTierThroughput,
			GlobalSecondaryIndexes: []schema.SecondaryIndex{
				schema.SecondaryIndex{
					IndexName: awsKey,
					KeySchema: []schema.KeySchema{
						{awsKey, schema.HashKey},
					},
					Projection:            schema.Projection{ProjectionType: schema.ProjectAll},
					ProvisionedThroughput: FreeTierThroughput,
				},
				schema.SecondaryIndex{
					IndexName: slackKey,
					KeySchema: []schema.KeySchema{
						{slackKey, schema.HashKey},
					},
					Projection:            schema.Projection{ProjectionType: schema.ProjectAll},
					ProvisionedThroughput: FreeTierThroughput,
				},
			},
		})
		if err != nil {
			return Client{}, fmt.Errorf("Failed to create table, %s", err)
		}
	}

	return Client{
		Dynamo: client,
	}, nil
}

// Client wraps the Dynago DynamoDB client.
type Client struct {
	Dynamo *dynago.Client
}
