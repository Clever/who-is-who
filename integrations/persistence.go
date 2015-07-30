package integrations

import (
	"fmt"
	"log"
	"math"
	"time"

	"gopkg.in/clever/kayvee-go.v2"
	"gopkg.in/underarmour/dynago.v1"
	"gopkg.in/underarmour/dynago.v1/schema"
)

const (
	cacheTTL = time.Minute * 10
)

// cacheList is meant to wrap a caching layer over a list users call to Dynamo.  This is used
// because Dynamo's throughput is allocated on a per record basis, therefore calls to /list run
// through ~100 items which is well over our single digit / second allocation (Dynamo watches
// requests on a 5 minute moving window).
// Given the nature of the data which is periodically synced, putting the list of users in a
// cache is reasonable.
type cacheList struct {
	Users       []User
	lastUpdated time.Time
}

// Client wraps the Dynago DynamoDB client. It also contains a cached copy of the list of users
// last returned by a database query.
type Client struct {
	Dynamo *dynago.Client
	Table  string
	cache  cacheList
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
	// return cached users if they were refreshed in the last 10 minutes
	if c.cache.lastUpdated.Add(cacheTTL).After(time.Now()) {
		return c.cache.Users, nil
	}

	res, err := c.Dynamo.Scan(c.Table).Execute()
	if err != nil {
		return []User{}, fmt.Errorf("Failed to scan table, %s", err)
	}

	users := make([]User, len(res.Items))
	for i, d := range res.Items {
		users[i] = UserFromDynago(d)
	}

	// update cache with user
	c.cache = cacheList{
		Users:       users,
		lastUpdated: time.Now(),
	}

	return users, nil
}

// NewClient creates a conection to DynamoDB, then creates the
func NewClient(tablename, endpoint, region, accessKey, secretKey string) (Client, error) {
	executor := dynago.NewAwsExecutor(endpoint, region, accessKey, secretKey)
	client := dynago.NewClient(executor)

	// DescribeTable 400's if table DNE
	_, err := client.DescribeTable(tablename)
	if err != nil {
		_, err := client.CreateTable(&schema.CreateRequest{
			TableName: tablename,
			AttributeDefinitions: []schema.AttributeDefinition{
				{emailKey, schema.String},
				{slackKey, schema.String},
				{awsKey, schema.String},
				{githubKey, schema.String},
			},
			KeySchema: []schema.KeySchema{
				{emailKey, schema.HashKey},
			},
			ProvisionedThroughput: FreeTierThroughput,
			GlobalSecondaryIndexes: []schema.SecondaryIndex{
				// index on AWS
				schema.SecondaryIndex{
					IndexName: awsIndex,
					KeySchema: []schema.KeySchema{
						{awsKey, schema.HashKey},
					},
					Projection:            schema.Projection{ProjectionType: schema.ProjectAll},
					ProvisionedThroughput: FreeTierThroughput,
				},
				// index on Slack
				schema.SecondaryIndex{
					IndexName: slackIndex,
					KeySchema: []schema.KeySchema{
						{slackKey, schema.HashKey},
					},
					Projection:            schema.Projection{ProjectionType: schema.ProjectAll},
					ProvisionedThroughput: FreeTierThroughput,
				},
				// index on github
				schema.SecondaryIndex{
					IndexName: githubIndex,
					KeySchema: []schema.KeySchema{
						{githubKey, schema.HashKey},
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
		Table:  tablename,
	}, nil
}
