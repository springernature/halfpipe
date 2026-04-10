#!/usr/bin/env bash

if [[ -f .halfpipe.io.yml ]]; then
    rm .halfpipe.io.yml
fi

"${HALFPIPE}" init > /dev/null

echo "Halfpipe"
diff --ignore-blank-lines .halfpipe.io.yml .halfpipe.io.expected.yml
