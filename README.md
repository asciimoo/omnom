# Omnom

A webpage bookmarking and snapshotting service.

Omnom is a rebooted implementation of @stef's original omnom project, big thanks for it.


## Requirements

go >= 1.14

## Setup & run

Checkout the repo and execute `go run omnom.go`

Settings can be configured in `config.yml` config file - restart webapp after updating.

## TODO

 - Add e-mail notification for login/token generation
 - Modernize js (proper use of async, etc..)
 - Package browser extension for firefox/chrome
 - Handle bookmark tags
 - Add paging to bookmark list view
 - Add profile view to create/remove addon-keys
 - Add style to the extension
