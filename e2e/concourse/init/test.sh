#!/usr/bin/env bash

if [[ -f .halfpipe.io ]]; then
    rm .halfpipe.io
fi

../../../halfpipe init > /dev/null

diff -w .halfpipe.io expected-halfpipe.yml
