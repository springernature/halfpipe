#!/usr/bin/env bash

if [[ -f actions.expected.yml ]]; then
  "${HALFPIPE}" -q -p actions -i .myCustomHalfpipePath.yml -o actions.actual.yml
  echo "  Actions"
  diff --ignore-blank-lines actions.actual.yml actions.expected.yml
fi

if [[ -f concourse.expected.yml ]]; then
  "${HALFPIPE}" -q -p concourse -i .myCustomHalfpipePath.yml -o concourse.actual.yml
  echo "  Concourse"
  diff --ignore-blank-lines concourse.actual.yml concourse.expected.yml
fi
