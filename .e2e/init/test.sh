#!/usr/bin/env bash

if [[ -f .halfpipe.io.yml ]]; then
    rm .halfpipe.io.yml
fi

../../halfpipe init > /dev/null

echo "  Halfpipe"
diff -w .halfpipe.io.yml .halfpipe.io.expected.yml && echo -e "\e[A✓"
