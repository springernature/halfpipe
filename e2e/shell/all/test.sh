#!/usr/bin/env bash

for f in $(ls | grep _expected.txt ); do
  taskName=$(echo $f | cut -d _ -f 1)
  echo "  task name: $taskName"
  ../../../halfpipe -q exec "$taskName" > "${f/expected/actual}"
  diff -w "$f" "${f/expected/actual}"
done
