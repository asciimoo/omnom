# Omnom

[![Translation status](https://hosted.weblate.org/widget/omnom/svg-badge.svg)](https://hosted.weblate.org/engage/omnom/)
[![GoDoc](https://godoc.org/github.com/asciimoo/omnom?status.svg)](https://pkg.go.dev/github.com/asciimoo/omnom)


A web content preservation service


 - Bookmark with website snapshots
 - Aggregate feeds
 - Follow ActivityPub streams

Check out our [wiki](https://github.com/asciimoo/omnom/wiki) for more information.


## Features

 - Self hosted
 - Bookmarking
 - RSS reader
 - [Fediverse/ActivityPub support](https://github.com/asciimoo/omnom/wiki/Fediverse-support)
 - Websites are captured as your browser renders it - saves the displayed content of dynamic pages as well
 - Multiple snapshots of the same URL with resource summary and compare/diff views
 - Web interface with multiuser support
 - Locally saved multimedia content
 - Flexible filtering - by date, free text search in content, tags, users, domains, URLs, etc..
 - [Documented API](https://github.com/asciimoo/omnom/wiki/API-documentation)


## Browser addon

Omnom browser addon is available for
- [Firefox](https://addons.mozilla.org/en-US/firefox/addon/omnom/)
- [Chrome/Chromium](https://chrome.google.com/webstore/detail/omnom/nhpakcgbfdhghjnilnbgofmaeecoojei)


## Installation

Single file binary release is available [here](https://github.com/asciimoo/omnom/releases/latest). (Don't forget to `chmod +x`)

Docker image is also available available, more details [here](https://github.com/asciimoo/omnom/wiki/Docker).

### Local build

go >= 1.24 required

 - Checkout the repo and execute `go get -u` in the root directory
 - Run `go build`

## Setup & run

 - Run `./omnom help` to list the available commands
 - Execute `./omnom listen` to start the web application


## Configuration

Settings can be configured in `config.yml` config file - don't forget to restart webapp after updating.

Execute `./omnom create-config config.yml` to generate a configuration file with the default configuration values.


## User handling

Omnom does not store passwords. Login requires one time login token, OAuth, or a remote user header.

Login tokens can be requested via email (this requires a valid SMTP configuration in `config.yml`) through the web interface or can be generated from command line using the `./omnom create-token [username] login`.

If you use Omnom behind a reverse proxy with authentication, you can pass the logged-in username in an HTTP header like `Remote-User` to automatically log in. Omnom can be configured to trust the header by setting the `remote_user_header` option in `config.yml`. Remote user header authentication can't be used with OAuth or open signups.


### Command line tool

Basic management actions are available using the command line tool (`go run omnom.go` or `go build; ./omnom`)

#### Available Commands
```
  create-bookmark      create new bookmark
  create-config        create default configuration file
  create-token         create new login/addon token for a user
  create-user          create new user
  diff-html            diff-html FILE1 FILE2
  generate-api-docs-md Generate Markdown API documentation
  help                 Help about any command
  listen               start server
  set-token            set new login/addon token for a user
  show-unread          show unread details
  show-user            show user details
  update-feeds         update RSS/Atom feeds
  validate-html        validate-html FILE
  completion           Generate the autocompletion script for the specified shell
```


## Translations

To contribute to localizations, please visit our [weblate](https://hosted.weblate.org/projects/omnom/)

[![Translation status](https://hosted.weblate.org/widget/omnom/multi-auto.svg)](https://hosted.weblate.org/engage/omnom/)



## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/asciimoo/omnom/issues) or join our [discord server](https://discord.gg/GAh4RCruh6)

## License

AGPLv3

## Funding

This project is funded through [NGI Zero Core](https://nlnet.nl/core), a fund established by [NLnet](https://nlnet.nl) with financial support from the European Commission's [Next Generation Internet](https://ngi.eu) program. Learn more at the [NLnet project page](https://nlnet.nl/project/Omnom-ActivityPub).

[<img src="https://nlnet.nl/logo/banner.png" alt="NLnet foundation logo" width="20%" />](https://nlnet.nl)
[<img src="https://nlnet.nl/image/logos/NGI0_tag.svg" alt="NGI Zero Logo" width="20%" />](https://nlnet.nl/core)
