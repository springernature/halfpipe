#!/usr/bin/env bash

for f in $(ls | grep .expected.txt ); do
  taskName=$(echo $f | cut -d . -f 1)
  echo "  task: $taskName"
  "${HALFPIPE}" -q exec "$taskName" > "${f/expected/actual}"
  diff --ignore-blank-lines "$f" "${f/expected/actual}"
done
