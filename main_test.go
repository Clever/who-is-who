package main

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Clever/kayvee-go"
	"github.com/Clever/who-is-who/integrations"
	"github.com/Clever/who-is-who/who-is-who"
	"github.com/gorilla/mux"
)

const (
	testTable = "test_users"
)

var (
	testDynamoEndpoint = requiredEnv("DYNAMO_ENDPOINT")
	router             *mux.Router

	testUserA = integrations.User{
		Email: "testA@test.com",
		Slack: "a",
		AWS:   "aAWS",
	}
	testUserB = integrations.User{
		Email: "testB@test.com",
		Slack: "b",
		AWS:   "bAWS",
	}
	testUserC = integrations.User{
		Email: "testC@test.com",
		Slack: "c",
		AWS:   "cAWS",
	}
	testUsers = map[string]integrations.User{
		testUserA.Email: testUserA,
		testUserB.Email: testUserB,
		testUserC.Email: testUserC,
	}
)

func setup(c integrations.Client) {
	err := c.SaveUsers(testUsers)
	if err != nil {
		log.Fatal(kayvee.FormatLog("who-is-who-testing", kayvee.Error, "batch insert error", m{
			"message": err.Error(),
		}))
	}
}

func teardown(c integrations.Client) {
	_, err := c.Dynamo.DeleteTable(testTable)
	if err != nil {
		log.Println(kayvee.FormatLog("who-is-who-testing", kayvee.Error, "delete table error", m{
			"message": err.Error(),
			"table":   testTable,
		}))
	}
}

func TestMain(m *testing.M) {
	c, err := integrations.NewClient(testTable, testDynamoEndpoint, "", "", "")
	if err != nil {
		log.Fatalf("Failed to connect to dynamoDB. Please run the local instance.")
	}
	router = hookUpRouter(dynamoConn{
		Dynamo: c,
	})

	setup(c)

	outcome := m.Run()
	defer os.Exit(outcome)

	teardown(c)
}

func TestListEndpoint(t *testing.T) {
	tester := httptest.NewServer(router)
	c := whoswho.NewClient(tester.URL)

	users, err := c.GetUserList()
	if err != nil {
		t.Fatalf("Failed AWS query: %s", err)
	} else if len(users) != 3 {
		t.Fatalf("wrong number of users returned")
	}

	for _, user := range users {
		u, valid := testUsers[user.Email]
		if !valid {
			t.Fatalf("returned user has invalid email: %s", user.Email)
		} else if u != user {
			t.Fatalf("returned user is missing info")
		}
	}
}

func TestAliasEndpointsSuccess(t *testing.T) {
	tester := httptest.NewServer(router)
	c := whoswho.NewClient(tester.URL)

	u, err := c.UserByEmail(testUserA.Email)
	if err != nil {
		t.Fatalf("Failed AWS query: %s", err)
	} else if u != testUserA {
		t.Fatalf("wrong user returned")
	}

	u, err = c.UserBySlack(testUserB.Slack)
	if err != nil {
		t.Fatalf("Failed AWS query: %s", err)
	} else if u != testUserB {
		t.Fatalf("wrong user returned")
	}

	u, err = c.UserByAWS(testUserC.AWS)
	if err != nil {
		t.Fatalf("Failed AWS query: %s", err)
	} else if u != testUserC {
		t.Fatalf("wrong user returned")
	}
}

func TestAliasEndpointsFailures(t *testing.T) {
	tester := httptest.NewServer(router)
	c := whoswho.NewClient(tester.URL)

	_, err := c.UserByEmail("nonexistant@email.com")
	if err == nil {
		t.Fatalf("No record should be returned")
	}
}
