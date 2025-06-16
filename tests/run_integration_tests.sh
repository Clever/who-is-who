#!/bin/sh

jar=dynamo-local/DynamoDBLocal.jar

# download the dynamo jar if necessary
if [ ! -e "$jar" ]
then
	if [ `uname` = "Darwin" ] ; then
		# this will prompt for java to be installed if necessary
		java -version
	else
    	sudo apt-get update && sudo apt-get install -y default-jre
    fi
    mkdir -p dynamo-local
    echo "Downloading dynamo server..."
    curl -L -k --url https://s3-us-west-2.amazonaws.com/dynamodb-local/dynamodb_local_latest.tar.gz -o dynamo-local/dynamodb_local_latest.tar.gz
    tar -zxvf dynamo-local/dynamodb_local_latest.tar.gz -C dynamo-local/
fi

# start up DynamoDBLocal for integration tests
java -jar "$jar" -sharedDb -inMemory -port 8002 &
sleep 5

# wait until port 8002 is listening, up to 15s
timeout=0
until nc -z localhost 8002 || [ $timeout -ge 15 ]; do
  echo "waiting for DynamoDB Local…"
  sleep 1
  timeout=$((timeout+1))
done

if ! nc -z localhost 8002; then
  echo >&2 "ERROR: DynamoDB Local did not start on port 8002"
  exit 1
fi

export AWS_ACCESS_KEY_ID=fake
export AWS_SECRET_ACCESS_KEY=fake
export AWS_REGION=us-west-1
export AWS_DYNAMO_REGION=us-west-1
export AWS_DYNAMO_ENDPOINT=http://localhost:8002

# run our tests
./node_modules/.bin/nodeunit ./tests/test-routes.js
err=$?

# kill all child processes to clean up
pkill -P $$
exit $err
