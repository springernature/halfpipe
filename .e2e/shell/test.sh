#!/usr/bin/env bash

for f in $(ls | grep .expected.txt ); do
  taskName=$(echo $f | cut -d . -f 1)
  echo "  task: $taskName"
  ${HALFPIPE} -q exec "$taskName" > "${f/expected/actual}"
  diff -w "$f" "${f/expected/actual}" && echo -e "\e[A✓"
done
