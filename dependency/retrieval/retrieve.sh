#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function main() {
  pushd "${PROGDIR}" > /dev/null
    go build retrieve.go
    ./retrieve "${@:-}"
    rm retrieve
  popd > /dev/null
}

main "${@:-}"
