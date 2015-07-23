#!/bin/sh

jar=DynamoDBLocal.jar

# download the dynamo jar if necessary
if [ ! -e "$jar" ]
then
	wget http://dynamodb-local.s3-website-us-west-2.amazonaws.com/dynamodb_local_latest.tar.gz
	tar -zxvf dynamodb_local_latest.tar.gz
fi

# start up DynamoDBLocal for integration tests
java -jar "$jar" -sharedDb -inMemory &
sleep 2
export DYNAMO_ENDPOINT=http://localhost:8000

# run our tests
go test -v ./...

# kill all child processes to clean up
pkill -P $$

