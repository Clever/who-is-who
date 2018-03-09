# who is who

A service that knits together user identities from various sources

Owned by infra

## API

- `/all` **(GET)** -- lists all items (same as `/alias` and `/list`)

- `/alias` **(GET)** -- lists all items (same as `/all` and `/list`)
- `/alias/:key` **(GET)** --
- `/alias/:key/:value` **(GET)** --
- `/alias/:key/:value` **(POST)** --
- `/alias/:key/:value/data/:path...` **(GET)** --
- `/alias/:key/:value/data/:path...` **(POST)** --

- `/alias/:key/:value/history/:path...`  **(GET)** --

- `/list` **(GET)** -- lists all items (same as `/all` and `/alias`)
- `/list/:key` **(GET)** --
- `/list/:key/:value` **(GET)** --
- `/list/:key/:value/data/:path...` **(GET)** --

