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

- `/alias/email/:email`
  - Returns info for a user with an email of `:email`.
- `/alias/slack/:handle`
  - Returns info for a user with a slack handle of `:handle`.
- `/alias/aws/:username`
  - Returns info for a user with an AWS username of `:username`.
- `/alias/github/:username`
  - Returns info for a user with a Github username of `:username`
- `/list`
  - Returns info for all users.
  - This list of users is cached for a period of 10 minutes.

Schema:

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


## Syncing

Please see [who-is-who-sync](https://github.com/Clever/who-is-who-sync).


## Storage

Data is stored in DynamoDB.
There are global secondary indexes in place on both the `slack` and `aws` attribute.


## Local development

Local development depends on having an instance of DynamoDB available.
It is suggested that you use the DynamoDBLocal jar for testing but using normal DynamoDB will work as well.
The database used needs to be populated with data for easy testing, it is suggested that you simply run `who-is-who-sync` to fill your database with real data.
At that point you can point your instance of `who-is-who` at `http://localhost:8000` (or whichever port you picked).


## Testing

The tests rely on access to a DynamoDB instance.
It is recommended that you use the [local DynamoDB instance](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Tools.DynamoDBLocal.html).


## Deployment

The following environment variables must be set to run the API:

- `DOMAIN` (this is used to filter the list of users in Slack as well as match Github usernames to users)
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `DYNAMO_TABLE`
- `DYNAMO_ENDPOINT`
- `DYNAMO_REGION`


## Changing Dependencies

### New Packages

When adding a new package, you can simply use `make vendor` to update your imports.
This should bring in the new dependency that was previously undeclared.
The change should be reflected in [Godeps.json](Godeps/Godeps.json) as well as [vendor/](vendor/).

### Existing Packages

First ensure that you have your desired version of the package checked out in your `$GOPATH`.

When to change the version of an existing package, you will need to use the godep tool.
You must specify the package with the `update` command, if you use multiple subpackages of a repo you will need to specify all of them.
So if you use package github.com/Clever/foo/a and github.com/Clever/foo/b, you will need to specify both a and b, not just foo.

```
# depending on github.com/Clever/foo
godep update github.com/Clever/foo

# depending on github.com/Clever/foo/a and github.com/Clever/foo/b
godep update github.com/Clever/foo/a github.com/Clever/foo/b
```

