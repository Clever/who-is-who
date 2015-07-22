package integrations

import (
	"fmt"
	"log"
	"math"

	"github.com/underarmour/dynago"
	"github.com/underarmour/dynago/schema"
	"gopkg.in/clever/kayvee-go.v2"
)

// Client wraps the Dynago DynamoDB client.
type Client struct {
	Dynamo *dynago.Client
	Table  string
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
		// properly figure out the indexes on the array: [i*batch, (i+1)*batch)
		firstIndex := i * batchLimit
		if firstIndex >= len(dynagoObjects) {
			break
		}
		lastIndex := int(math.Min(float64((i+1)*batchLimit), float64(len(dynagoObjects))))

		// perform write and check for unprocessed items
		res, err := c.Dynamo.BatchWrite().Put(c.Table, dynagoObjects[firstIndex:lastIndex]...).Execute()
		if err != nil {
			return fmt.Errorf("Error while executing batch write: %s", err)
		} else if failedPuts := res.UnprocessedItems.GetPuts(c.Table); len(failedPuts) > 0 {
			// it is unlikely that we will have failed writes so we simply print it
			// for use in debugging if that case does happen.
			log.Println(kayvee.FormatLog("who-is-who", kayvee.Error, "batchWrite fails", map[string]interface{}{
				"num":    len(failedPuts),
				"values": failedPuts,
			}))
		}
	}

	return nil
}

// GetUser crafts a query for a single user based on the specified index and user information.
func (c Client) GetUser(idx Index, value string) (User, error) {
	res, err := c.Dynamo.
		Query(c.Table).
		IndexName(idx.Index).
		KeyConditionExpression(fmt.Sprintf("%s = :username", idx.Field), dynago.Param{Key: ":username", Value: value}).
		Execute()

	if err != nil {
		return User{}, fmt.Errorf("Failed to make query with '%s'=='%s' due to: %s", idx.Field, value, err)
	} else if res.Count == 0 {
		return User{}, ErrUserDNE
	}

	return UserFromDynago(res.Items[0]), nil
}

// GetUserList returns all users.
func (c Client) GetUserList() ([]User, error) {
	res, err := c.Dynamo.Scan(c.Table).Execute()
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
func NewClient(table, endpoint, region, accessKey, secretKey string) (Client, error) {
	executor := dynago.NewAwsExecutor(endpoint, region, accessKey, secretKey)
	client := dynago.NewClient(executor)

	// DescribeTable 400's if table DNE
	_, err := client.DescribeTable(table)
	if err != nil {
		_, err := client.CreateTable(&schema.CreateRequest{
			TableName: table,
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
		Table:  table,
	}, nil
}