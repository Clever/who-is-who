package server

import (
	"context"

	"github.com/Clever/who-is-who/gen-go/models"
)

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=mock_controller.go -package=server

// Controller defines the interface for the who-is-who service.
type Controller interface {

	// HealthCheck makes a GET request to /_health.
	// Checks if the service is healthy
	HealthCheck(ctx context.Context) error

	// GetUserByAlias makes a GET request to /alias/{alias_type}/{alias_value}.
	// Returns info for a user based on an alias
	GetUserByAlias(ctx context.Context, i *models.GetUserByAliasInput) (*models.User, error)

	// List makes a GET request to /list.
	// Returns info for all user
	List(ctx context.Context) (*models.UserList, error)
}
