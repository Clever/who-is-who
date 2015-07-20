package cleveraws

import (
	"strings"

	"github.com/Clever/who-is-who/integrations"
	"github.com/underarmour/dynago/schema"
)

const (
	key = "aws"
)

var (
	Index       = integrations.Index{"aws", "aws"}
	DynamoIndex = schema.SecondaryIndex{
		IndexName: Index.Index,
		KeySchema: []schema.KeySchema{
			{key, schema.HashKey},
		},
		Projection:            schema.Projection{ProjectionType: schema.ProjectAll},
		ProvisionedThroughput: integrations.FreeTierThroughput,
	}
)

func init() {
	integrations.GlobalSecondaryIndexes = append(integrations.GlobalSecondaryIndexes, DynamoIndex)
}

// AwsService does the computation to form AWS usernames with a first initial and last name.
type AwsService struct{}

// Init implements the InfoSource interface. There are no API calls made.
func (a AwsService) Init(_ string) error { return nil }

// Fill uses the first and last name to form an AWS username.
func (a AwsService) Fill(m integrations.UserMap) integrations.UserMap {
	for email, user := range m {
		if user.FirstName != "" && user.LastName != "" {
			user.AWS = strings.ToLower(user.FirstName[0:1] + user.LastName)
		}
		m[email] = user
	}
	return m
}
