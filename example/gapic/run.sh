#!/bin/bash

set -o pipefail
set -u

declare -r g_pwd="$(realpath `dirname -- "${BASH_SOURCE[0]}"` )"
declare -r g_bin="$(realpath ${g_pwd}/../..)"

pushd "${g_pwd}"

mkdir -p ./gen && \
PATH="${g_bin}:${PATH}" protoc \
    --go_out=:./gen \
    --go_opt='module=go.example.com/apis' \
    --go_gapic_out=:./gen \
    --go_gapic_opt='go-gapic-package=go.example.com/apis/gapic;gapic' \
    --go_gapic_opt='module=go.example.com/apis' \
    --go_gapic_opt='omit-snippets' \
    proto/*.proto \
    -I. -I proto

popd

