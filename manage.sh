#!/bin/sh

BASE_DIR="$(dirname -- "`readlink -f -- "$0"`")"
OMNOM_BASE_URL="http://127.0.0.1:7332/"
CONFIG_PATH="$BASE_DIR/tests/test_config.yml"
ACTION="$1"
[ -z "$ACTION" ] || shift

CHROME_EXT_ZIP='omnom_ext_chrome.zip'
FF_EXT_ZIP='omnom_ext_firefox.zip'
EXT_SRC_ZIP='omnom_ext_src.zip'

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
build_test_addon     - Build test addon
build_addon_artifact - Build addon artifacts to distribute to addon stores

Other
-----
sync_translations    - Synchronize translations from weblate
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
    echo "[!] Warning: The default manifest.json is for chrome browsers, overwrite it with manifest_ff.json for firefox"
    cd ext
    npm run build
    cd ..
}

build_test_addon() {
    cd ext
    npm run build-test
    cd ..
}

sync_translations() {
    DIR="/tmp/tr"
    mkdir $DIR
    cd $DIR
    wget "https://hosted.weblate.org/download/omnom/?format=zip" -O tr.zip
    unzip tr.zip
    cd -
    cp "$DIR/omnom/app/localization/locales/"*json localization/locales/
    git add localization/locales/
    git commit localization/locales/ -m '[enh] synchronize translations from weblate'
    rm -r "$DIR"
}

build_addon_artifact() {
    build_addon
    [ -e "$EXT_SRC_ZIP" ] && rm "$EXT_SRC_ZIP" || :
    [ -e "$EXT_ZIP" ] && rm "$EXT_ZIP" || :
    cd ext
    zip -r "../$EXT_SRC_ZIP" README.md src utils package* webpack.config.js
    cd build
    zip "../../$CHROME_EXT_ZIP" ./* icons/* -x manifest_ff.json
    cp manifest.json /tmp
    cp manifest_ff.json manifest.json
    zip "../../$FF_EXT_ZIP" ./* icons/* -x manifest_ff.json
    mv /tmp/manifest.json manifest.json
    cd ../../
}

[ "$(command -V "$ACTION" | grep ' function$')" = "" ] \
	&& help "action not found" \
	|| "$ACTION" "$@"
