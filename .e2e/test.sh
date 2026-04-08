#!/usr/bin/env bash
set -e

E2E_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HALFPIPE="$(cd "${E2E_DIR}/.." && pwd)/halfpipe"
HALFPIPE_BRANCH=main

runTest() {
  dir=${1%*/}
  echo "* Running ${dir##*/}"

  cd "${dir}"
  if [[ -f test.sh ]]; then
    ./test.sh
    cd - > /dev/null
    return
  fi

  if [[ -f actions.expected.yml ]]; then
    "${HALFPIPE}" -q -p actions -o actions.actual.yml
    echo "  Actions"
    diff --ignore-blank-lines actions.actual.yml actions.expected.yml
  fi

  if [[ -f concourse.expected.yml ]]; then
    "${HALFPIPE}" -q -p concourse -o concourse.actual.yml
    echo "  Concourse"
    diff --ignore-blank-lines concourse.actual.yml concourse.expected.yml
    if command -v fly > /dev/null; then
      fly validate-pipeline -c concourse.actual.yml &> /dev/null
    fi
  fi

  cd - > /dev/null
}

export HALFPIPE
export HALFPIPE_BRANCH
export -f runTest

if command -v parallel > /dev/null; then
  ls -d "${E2E_DIR}"/*/  | parallel -k -j16 runTest
else
  ls -d "${E2E_DIR}"/*/ | xargs -I{} -P1 bash -c 'runTest "$@"' _ {}
fi
