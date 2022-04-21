#!/bin/sh

BASE_DIR="$(dirname -- "`readlink -f -- "$0"`")"
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
run_e2e_tests        - Run browser tests

Build
-----
build_css            - Build css files
build_addon          - Build addon
"
	[ -z "$1" ] && exit 0 || exit 1
}


install_js_deps() {
    cd sass
    npm i
    cd ..
    cd ext
    npm i
    cd ..
}

install_test_deps() {
    cd tests/e2e/extension
    npm i
    cd "$BASE_DIR"
}

run_e2e_tests() {
    cd tests/e2e/extension
    nodejs test.js
    cd "$BASE_DIR"
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
