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

Binaries for our officially supported platforms can be found on the [releases page](https://github.com/PagerDuty/go-pdagent/releases).

## Usage

On first run we recommend running `pdagent init` to generate a default config file. By default this file will live in `~/.pdagent` along with any other artifacts.

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

## Architecture

![pdagent architecture diagram](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.github.com/PagerDuty/go-pdagent/main/docs/architecture-diagram.txt)

At a high level, the agent has three key components:

- Server: The daemon itself where most of the heavy lifting occurs.
- Client: An HTTP client to simplify making requests against the server.
- CLI: A command line tool for working with both the server and client commands.

## Development

Looking to contribute? See [development](/docs/development.md) for some helpful tips.

## Current Status

This project aims to eventually replace the existing `pdagent` project, but with some goals in mind before doing so:

- [x] Events API V1 support.
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
    - [x] `pd-nagios`
    - [ ] `pd-sensu`
    - [ ] `pd-zabbix`
