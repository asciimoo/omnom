# Omnom

A webpage bookmarking and snapshotting service.

Omnom consists of two parts; a multi-user web application that accepts bookmarks and snapshots and a browser extension responsible for bookmark and snapshot creation.


Omnom is a rebooted implementation of @stef's original omnom project, big thanks for it.


## Requirements

go >= 1.20

## Setup & run

 - Checkout the repo and execute `go get -u`
 - Copy `config.yml_sample` to `config.yml`
 - Execute `go run omnom.go listen` in the repo root

Settings can be configured in `config.yml` config file - don't forget to restart webapp after updating.

### Command line tool

Basic management actions are available using the command line tool (`go run omnom.go` or `go build; ./omnom`)

#### Available Commands
```
  create-token         create new login/addon token for a user
  create-user          create new user
  generate-api-docs-md Generate Markdown API documentation
  help                 Help about any command
  listen               start server
  set-token            set new login/addon token for a user
  show-user            show user details
  completion           Generate the autocompletion script for the specified shell
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
