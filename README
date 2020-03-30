# Pagerduty Agent

An agent daemon for to aid in creating PagerDuty events.

Goals of this project include providing:

- A command-line interface for creating PagerDuty events.
- A local entry point for PagerDuty's Events API.
- Ensuring that events are properly ordered for each integration.
- Handling backpressure or when PagerDuty is inaccessible.

## Installation

Currently the agent needs to be built from source, but releasing pre-built binaries and distributing through common package managers is on our roadmap.

For the time being:

- Install Go: https://golang.org/doc/install#install
- Clone the project: https://github.com/PagerDuty/pagerduty-agent
- Run `make build`

You should now have a working `pagerduty-agent` binary.

## Usage

On first run we recommend running `pagerduty-agent init` to generate a default config file. By default this is created as `~/.pagerduty-agent.yaml` and includes options such as which address to run the server on, the client/server secret, and where the database should live.

Once the config has been created, to start the daemon:

```
pagerduty-agent server
```

There are a number of other commands available that are listed as part of the command's help command:

```
pagerduty-agent help
```

Perhaps the most common command, sending events:

```
pagerduty-agent send \
  -k v4g5q7yie1qee1yio2uimz8yfbjenvj9 \
  -t trigger \
  -d "This is a test event" \
  -c "PagerDuty Test" \
  -u "http://pagerduty.com" \
  -f some_field=some_value
```

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

**Additional publishing details TBD during migration.**

## Architecture

At a high level, the agent has three key components:

- Server: The daemon itself where most of the heavy lifting occurs.
- Client: An HTTP client to simplify making requests against the server.
- CLI: A command line tool for working with both the server and client commands.

The server leverages several packages found under `pkg`, described below.

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


## Troubleshooting

WIP

## Contributing

WIP