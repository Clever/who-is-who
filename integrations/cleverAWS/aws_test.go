package cleveraws

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Clever/who-is-who/integrations"
)

func TestAWS(t *testing.T) {
	assert := assert.New(t)

	userMap := integrations.UserMap{
		"who@car.es": integrations.User{
			FirstName: "First Name",
			LastName:  "Last Name",
		},
	}
	service := AwsService{}

	userMap, err := service.Fill(userMap)
	assert.NoError(err)
	assert.Equal(1, len(userMap))

	user, ok := userMap["who@car.es"]
	assert.True(ok)
	assert.Equal("flastname", user.AWS)
}
