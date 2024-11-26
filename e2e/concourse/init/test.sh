#!/usr/bin/env bash

if [[ -f .halfpipe.io.yml ]]; then
    rm .halfpipe.io.yml
fi

../../../halfpipe init > /dev/null

diff -w .halfpipe.io.yml expected-halfpipe.yml
