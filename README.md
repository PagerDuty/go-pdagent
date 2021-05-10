**Note: This project is currently in beta, for the current Python-based agent see: https://github.com/PagerDuty/pdagent**

# PagerDuty Agent

An agent daemon to aid in creating PagerDuty events.

Goals of this project include providing:

- A command-line interface for creating PagerDuty events.
- A local entry point for PagerDuty's Events API.
- Ensuring that events are properly ordered for each integration.
- Handling back pressure or when PagerDuty is inaccessible.

If you're looking for a more comprehensive PagerDuty API Go client library and CLI, see: https://github.com/PagerDuty/go-pagerduty

## Installation

Currently the agent needs to be built from source, but releasing pre-built binaries and distributing through common package managers is on our roadmap.

For the time being:

- Install Go: https://golang.org/doc/install#install
- Clone the project: https://github.com/PagerDuty/go-pdagent
- Run `make build`

You should now have a working `pdagent` binary.

## Usage

On first run we recommend running `pdagent init` to generate a default config file. By default during local development this file will live in `~/.pdagent` along with any other artifacts.

Once the config has been created, to start the daemon:

```
pdagent server
```

There are a number of other commands available that are listed as part of the command's help command:

```
pdagent help
```

Perhaps the most common command, sending events:

```
pdagent enqueue \
  -k your_key_goes_here \
  -t trigger \
  -d "This is only a test" \
  -u "http://pagerduty.com" \
  -e "error" \
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

## Architecture

![pdagent architecture diagram](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.github.com/rafusel/go-pdagent/add-architecture-diagram/docs/architecture-diagram.txt)

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

## Current Status

This project aims to eventually replace the existing `pdagent` project, but with some goals in mind before doing so:

- [ ] Events API V1 support.
- [X] Events API V2 support.
    - [X] Parity with existing `pd-send` functionality.
    - [X] Comprehensive Events API V2 payload support (no links / images yet).
- [ ] HTTP configuration.
    - [ ] Custom cert files.
    - [x] Proxy and firewall support.
    - [ ] Local server security.
- [X] Event queuing.
- [X] Persistent queuing.
- [x] Legacy command wrappers.
    - [x] `pd-send`
    - [x] `pd-queue`
- [ ] Releasing
    - [x] Init and pre/post install scripts.
    - [X] Github release support.
        - [X] Source
        - [X] Darwin
        - [X] Linux (deb/rpm)
        - [X] Checksums.
        - [X] Signature files.
    - [ ] `deb` repo support.
    - [ ] `rpm` repo support.
        - [ ] Signed packages.
- [ ] `pdagent-integrations` support.
    - [ ] `pd-nagios`
    - [ ] `pd-sensu`
    - [ ] `pd-zabbix`
