# who is who

A service that knits together user identities from various sources.

Owned by infra

## API

The header `X-WIW-AUTHOR` is required for all requests. Please set it to your email address.

- `/all` **(GET)** -- lists all items (same as `/alias` and `/list`)

- `/alias` **(GET)** -- lists all items (same as `/all` and `/list`)
- `/alias/:key` **(GET)** -- lists all users who have any value set for `key` (same as `/list/:key`)
- `/alias/:key/:value` **(GET)** -- If there is exactly one user matching `{key:value}`, return that user. Otherwise, error. (For 0 or more than 1 user, use `/list/:key/:value`).
- `/alias/:key/:value` **(POST)** -- Requires a JSON object in the body and exactly one user matching `{key:value}`. Sets each key present in body to the value provided in body, leaving any other keys alone.
- `/alias/:key/:value/data/:path...` **(GET)** -- Gets the value of a single key for the single matching user. For example, `/alias/key/value/data/key2/innerkey` returns the value of `user.key2.innerykey`.
- `/alias/:key/:value/data/:path...` **(POST)** -- Requires a body. Sets the value of matching user's path to an object in body. This one is somewhat weird, you're probably better off using `POST /alias/:key/:value`.

- `/alias/:key/:value/history/:path...`  **(GET)** -- Gets the full history of the given path for the single matching user.

- `/list` **(GET)** -- lists all items (same as `/all` and `/alias`)
- `/list/:key` **(GET)** -- lists all users who have any value set for `key` (same as `/alias/:key`)
- `/list/:key/:value` **(GET)** -- returns an array (possibly empty) of all users matching `{key:value}`.
- `/list/:key/:value/data/:path...` **(GET)** -- returns an array consisting of the value of path for each matching user. For example, `/list/active/true/data/email` returns an array of the emails of all active users.

## Schema

`who-is-who` is backed by three DynamoDB tables in us-west-1. While these can be edited by hand via AWS console or CLI, please be very careful. The full specs can be found [here](./storage/dynamo.js).

- `whoswho-objects` is a full collection of each user along with arbitrary key-value data about them. It uses a UUID field named `_whoid` as the primary key. Email is required field, although this is not enforced on the database level.
- `whoswho-paths` serves sort-of like a very general index over `whoswho-objects`. It uses a composite primary key consisting of a partition key `path` for the field and a sort key consisting of the value of that field for a particular user, then the null character `\u0000`, then that user's `_whoid`. Each item also has `_whoid` as a separate field.
For example, for a user in `whoswho-objects`
`{ "_whoid": "ID", "email": "user.name@example.com", "slack": "user"}`
if everything is in sync, there would be the following items in `whoswho-paths`
``` json
{ "path": "email", "val_whoid": "user.name@example.com\u0000ID", "_whoid": "ID"}
{ "path": "slack", "val_whoid": "user\u0000ID", "_whoid": "ID"}
```
- `whoswho-history` keeps a change log.

## Debugging

If there's an issue, it's likely that `whoswho-objects` and `whoswho-paths` are out of sync. This likely needs to be fixed by manually changing the tables. You can use the AWS DynamoDB CLI to put the `\u0000` into `whoswho-paths`.

There are no API endpoints at the moment for deleting keys, but if you set the value of a key to the empty string (e.g. using `POST /alias/:key/:value`, it will be deleted. The preferred way to delete a user is to set their `active` to `false`.
