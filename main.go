package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"gopkg.in/Clever/pathio.v3"

	models "github.com/Clever/who-is-who/gen-go/models"
	"github.com/Clever/who-is-who/gen-go/server"
)

// Controller implements server.Controller
type Controller struct {
	Users models.UserList
}

// HealthCheck returns an error if the application is unhealthy
func (c *Controller) HealthCheck(ctx context.Context) error {
	return nil
}

// GetUserByAlias looks up a user according to an alias (ex: alias_type='email',alias_value='some.user@clever.com')
func (c *Controller) GetUserByAlias(ctx context.Context, i *models.GetUserByAliasInput) (*models.User, error) {
	if i == nil {
		return nil, fmt.Errorf("invalid input")
	}

	switch i.AliasType {
	// Valid alias types
	case "slack":
	case "email":
	case "github":
	case "aws":
		break
	default:
		return nil, fmt.Errorf("invalid alias type")
	}
	if i.AliasValue == "" {
		return nil, fmt.Errorf("alias value cannot be empty-string")
	}

	for _, user := range c.Users {
		if (i.AliasType == "slack" && *user.SLACK == i.AliasValue) ||
			(i.AliasType == "email" && *user.Email == i.AliasValue) ||
			(i.AliasType == "github" && *user.Github == i.AliasValue) ||
			(i.AliasType == "aws" && *user.Aws == i.AliasValue) {
			return user, nil
		}
	}
	return nil, fmt.Errorf("No matching user found")
}

// List returns all users
func (c *Controller) List(ctx context.Context) (*models.UserList, error) {
	return &c.Users, nil
}

// loadUsers reads a JSON from a local file path or S3 path
func loadUsers(path string) (models.UserList, error) {
	reader, err := pathio.Reader(path)
	if err != nil {
		return models.UserList{}, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	var c models.UserList
	err = json.Unmarshal(buf.Bytes(), &c)
	if err != nil {
		return models.UserList{}, err
	}
	return c, nil
}

func main() {
	addr := flag.String("addr", ":8080", "Address to listen at")
	path := flag.String("users", "./users.json", "Path to file (local or S3) containing users")
	flag.Parse()

	users, err := loadUsers(*path)
	if err != nil {
		log.Fatal("Unable to load users:", err.Error())
	}

	controller := Controller{Users: users}
	s := server.New(&controller, *addr)
	// Serve should not return
	log.Fatal(s.Serve())
}
