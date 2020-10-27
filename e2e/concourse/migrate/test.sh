#!/usr/bin/env bash

if [[ -f .halfpipe.io ]]; then
    rm .halfpipe.io
fi

cp .halfpipe.io-test .halfpipe.io

../../../halfpipe migrate > /dev/null

diff -w .halfpipe.io expected-halfpipe.yml

