#!/usr/bin/env bash

export HALFPIPE_BRANCH="feature-a/branch-b"

if [[ -f actions.expected.yml ]]; then
  "${HALFPIPE}" -q -p actions -o actions.actual.yml
  echo "  Actions"
  diff --ignore-blank-lines actions.actual.yml actions.expected.yml
fi

if [[ -f concourse.expected.yml ]]; then
  "${HALFPIPE}" -q -p concourse -o concourse.actual.yml
  echo "  Concourse"
  diff --ignore-blank-lines concourse.actual.yml concourse.expected.yml
fi
