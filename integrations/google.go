package integrations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	directory "google.golang.org/api/admin/directory_v1"
)

const (
	publicDataView = "DOMAIN_PUBLIC"
	domain         = "adicu.com"
)

// GetGoogleCredentialsFromJSON loads a JSON containing the Google Apps credentials from the
// provided filepath.
func GetGoogleCredentialsFromJSON(filepath string) (GoogleCredentials, error) {
	var creds GoogleCredentials

	serviceAccountJSON, err := ioutil.ReadFile(filepath)
	if err != nil {
		return creds, fmt.Errorf("Could not read service account credentials file, %s => {%s}", filepath, err)
	}

	err = json.Unmarshal(serviceAccountJSON, &creds)
	if err != nil {
		return creds, fmt.Errorf("Could not parse service account credentials file, %s => {%s}", filepath, err)
	}

	return creds, nil
}

// GetGoogleCredentialsFromEnv builds a google config from environment variables.
func GetGoogleCredentialsFromEnv() (GoogleCredentials, error) {
	var creds GoogleCredentials

	return creds, nil
}

// GetGoogleClient takes a Google credentials struct and connects to Google's Oauth2 services
// to authenticate a client for their admin directory service.
func GetGoogleClient(creds GoogleCredentials) (*directory.Service, error) {
	serviceAccountJSON, err := json.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal Google Service Account credentials => {%s}", err)
	}
	log.Println(string(serviceAccountJSON))

	config, err := google.JWTConfigFromJSON(serviceAccountJSON,
		directory.AdminDirectoryUserScope,
		directory.AdminDirectoryUserReadonlyScope,
	)

	client, err := directory.New(config.Client(oauth2.NoContext))
	if err != nil {
		return nil, fmt.Errorf("Could not create directory service client => {%s}", err)
	}

	// TODO: remove this test
	users, err := client.Users.List().ViewType(publicDataView).Domain(domain).Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to query all users => {%s}", err)
	}
	for _, u := range users.Users {
		fmt.Println(u.Name.FullName)
	}

	return client, nil
}

// GoogleCredentials represents the information needed for a Google Apps service account
// authentication.
type GoogleCredentials struct {
	ClientEmail  string `json:"client_email"`
	ClientID     string `json:"client_id"`
	PrivateKey   string `json:"private_key"`
	PrivateKeyID string `json:"private_key_id"`
	Type         string `json:"type"`
}
