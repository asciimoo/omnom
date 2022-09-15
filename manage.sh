#!/bin/sh

BASE_DIR="$(dirname -- "`readlink -f -- "$0"`")"
OMNOM_BASE_URL="http://127.0.0.1:7332/"
CONFIG_PATH="$BASE_DIR/tests/test_config.yml"
ACTION="$1"
[ -z "$ACTION" ] || shift

cd -- "$BASE_DIR"
set -e


help() {
	[ -z "$1" ] || printf 'Error: %s\n' "$1"
	echo "Omnom manage.sh help

Commands
========
help                 - Display help

Dependencies
------------------
install_js_deps      - Install or install frontend dependencies (required only for development)
install_test_deps    - Install or install test dependencies

Tests
-----
run_unit_tests       - Run unit tests
run_e2e_tests        - Run browser tests
start_test_server    - Launch test server

Build
-----
build_css            - Build css files
build_addon          - Build addon

========

Execute 'go run omnom.go' or 'go build && ./omnom' for application related actions
"
	[ -z "$1" ] && exit 0 || exit 1
}


install_js_deps() {
    cd sass
    npm i --unsafe-perm -g
    cd ..
    cd ext
    npm install cross-env
    npm i
    cd ..
}

install_test_deps() {
    cd tests/e2e/extension
    npm i
    cd "$BASE_DIR"
}

run_unit_tests() {
    go test ./...
}

run_e2e_tests() {
    go run omnom.go --config "$CONFIG_PATH" create-user test test@127.0.0.1 || :
    go run omnom.go --config "$CONFIG_PATH" set-token test login 0000000000000000000000000000000000000000000000000000000000000000
    go run omnom.go --config "$CONFIG_PATH" set-token test addon 0000000000000000000000000000000000000000000000000000000000000000
    cd tests/e2e/extension
    node test.js "$OMNOM_BASE_URL"
    cd "$BASE_DIR"
}

start_test_server() {
    go run omnom.go --config "$CONFIG_PATH" listen
}

build_css() {
    cd sass
    npm run build:css
    cd ..
}

build_addon() {
    cd ext
    npm run build
    cd ..
}


[ "$(command -V "$ACTION" | grep ' function$')" = "" ] \
	&& help "action not found" \
	|| "$ACTION" "$@"
