#!/usr/bin/env bash

for f in `find . -name '*_expected.txt'`; do
  taskName="${f:2:(-13)}"
  echo "  task name: $taskName"
  ../../../halfpipe -q exec "$taskName" > "${f/expected/actual}"
  diff -w "$f" "${f/expected/actual}"
done
