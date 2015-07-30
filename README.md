# Who's Who?

Given all the tools we use in development, notifications often lose meaning.
The more targeted a notification, the more useful it is to the worker.

Who's who is a service that aggregates aliases from services to provide an API directory.
It is intended to be used as infrastructure for internal tools to more easily provide targeted emails and slack notifications.

Slack profile info is treated as the base source of truth when refreshing the directory.

## Lookups

- Email
- Slack Username
- Aws (Clever's scheme of first initial + last name)
- Github username

## API

schema:

```js
{
  "first_name": "abc",
  "last_name": "def"
  "email": "abc@test.com",
  "slack": "abc",
  "phone": "123-456-7890",
  "aws": "adef",
  "github": "abc"
}
```

- `/alias/email/:email`
  - Returns info for a user with an email of `:email`
- `/alias/slack/:handle`
  - Returns info for a user with a slack handle of `:handle`
- `/alias/aws/:username`
  - Returns info for a user with an AWS username of `:username`
- `/alias/github/:username`
  - Returns info for a user with a Github username of `:username`
- `/list`
  - Returns info for all users


## Syncing

Please see [who-is-who-sync](https://github.com/Clever/who-is-who-sync).


## Storage

Data is stored in DynamoDB.

## Testing

The tests rely on access to a DynamoDB instance.
It is recommended that you use the [local DynamoDB instance](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Tools.DynamoDBLocal.html).

## Deployment

The following environment variables must be set to run the API:

- `DOMAIN`
- `PORT`
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `DYNAMO_TABLE`
- `DYNAMO_ENDPOINT`
- `DYNAMO_REGION`

