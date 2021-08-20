# Development

## Building From Source

- Install Go: https://golang.org/doc/install#install
- Clone the project: https://github.com/PagerDuty/go-pdagent
- Run `make build`

You should now have a working `pdagent` binary.

## Releasing

For local builds and releases, install GoReleaser: https://goreleaser.com/

Builds are signed with GPG. To test your configuration, run the following:

```
gpg --list-secret
```

If nothing was output, run the following (for our purposes we recommend leaving the passphrase blank):

```
gpg --full-generate-key
```

Once GPG is set up, there are two commands for building:

```bash
make release # Regular distributable release, with publishing.
make release-test # To build a local snapshot release without publishing.
```

## Packages

The agent leverages several packages found under `pkg`, described below.

### `persistentqueue`

A database-backed event queue, used directly by the daemon server.

Events are added to the database as they're enqueued, updated after a response is received from PagerDuty, and used during startup to check for any unsent events. It also powers various operational commands (like `status` and `retry`).

Most of the actual queuing is handled by the `eventqueue` package that `persistentqueue` lleverages.

### `eventqueue`

The event queue responsible for ensuring ordering and handling backpressure, used by `persistentqueue`.

The event queue maintains a worker and buffered channel for every routing key that it's aware of, effectively the individual queues for each integration. It's designed to be used asynchronously with responses communicated over response channels.

Events are represented as "jobs" and processed by "processors," currently an event processor backed by `eventsapi`.

### `eventsapi`

A small helper library used for sending events to both Events API V1 and V2 endpoints. Currently this package is leveraged by `eventqueue` when processing events.
