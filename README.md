# Omnom

A webpage bookmarking and snapshotting service.

Omnom consists of two parts; a multi-user web application that accepts bookmarks and snapshots and a browser extension responsible for bookmark and snapshot creation.


Omnom is a rebooted implementation of @stef's original omnom project, big thanks for it.


## Requirements

go >= 1.14

## Setup & run

Checkout the repo and execute `go get -u` then `go run omnom.go listen` in the repo root.

Settings can be configured in `config.yml` config file - restart webapp after updating.

### Command line tool

Basic management actions are available using the command line tool (`go run omnom.go` or `go build; ./omnom`)

#### Available Commands
```
  create-token create new login token for a user
  create-user  create new user
  listen       start server
  show-user    show user details
  help         Help about any command
  completion   generate the autocompletion script for the specified shell
```

### Browser addon

Omnom browser addon is available for
- [Firefox](https://addons.mozilla.org/en-US/firefox/addon/omnom/)
- [Chrome/Chromium](https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei)

## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/asciimoo/omnom/issues) or join our [discord server](https://discord.gg/GAh4RCruh6)

## License

AGPLv3

## Funding

This project is funded through [NGI Zero Core](https://nlnet.nl/core), a fund established by [NLnet](https://nlnet.nl) with financial support from the European Commission's [Next Generation Internet](https://ngi.eu) program. Learn more at the [NLnet project page](https://nlnet.nl/project/Omnom-ActivityPub).

[<img src="https://nlnet.nl/logo/banner.png" alt="NLnet foundation logo" width="20%" />](https://nlnet.nl)
[<img src="https://nlnet.nl/image/logos/NGI0_tag.svg" alt="NGI Zero Logo" width="20%" />](https://nlnet.nl/core)
