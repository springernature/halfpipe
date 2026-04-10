#!/usr/bin/env bash
set -uo pipefail

E2E_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HALFPIPE="$(cd "${E2E_DIR}/.." && pwd)/halfpipe"
HALFPIPE_BRANCH=main
RESULTS_DIR=$(mktemp -d)

# Use colored diff when outputting to a terminal, plain diff otherwise
if [[ -t 1 ]] && diff --color /dev/null /dev/null 2>/dev/null; then
  DIFF_COLOR="--color"
else
  DIFF_COLOR=""
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
RESET='\033[0m'

if [[ ! -t 1 ]]; then
  RED=""
  GREEN=""
  BOLD=""
  RESET=""
fi

runTest() {
  dir=${1%*/}
  name=${dir##*/}
  output=""
  failed=0

  cd "${dir}"

  if [[ -f test.sh ]]; then
    if ! test_output=$(./test.sh 2>&1); then
      failed=1
      output+=$(printf "\n${RED}${BOLD}FAIL: %s${RESET}\n%s\n" "${name}" "${test_output}")
    fi
    cd - > /dev/null
  else
    if [[ -f actions.expected.yml ]]; then
      if ! "${HALFPIPE}" -q -p actions -o actions.actual.yml 2>/dev/null; then
        failed=1
        output+=$(printf "\n${RED}${BOLD}FAIL: %s / Actions${RESET}\n  halfpipe command failed\n" "${name}")
      elif ! diff_output=$(diff --ignore-blank-lines ${DIFF_COLOR} actions.actual.yml actions.expected.yml 2>&1); then
        failed=1
        output+=$(printf "\n${RED}${BOLD}FAIL: %s / Actions${RESET}\n%s\n" "${name}" "${diff_output}")
      fi
    fi

    if [[ -f concourse.expected.yml ]]; then
      if ! "${HALFPIPE}" -q -p concourse -o concourse.actual.yml 2>/dev/null; then
        failed=1
        output+=$(printf "\n${RED}${BOLD}FAIL: %s / Concourse${RESET}\n  halfpipe command failed\n" "${name}")
      elif ! diff_output=$(diff --ignore-blank-lines ${DIFF_COLOR} concourse.actual.yml concourse.expected.yml 2>&1); then
        failed=1
        output+=$(printf "\n${RED}${BOLD}FAIL: %s / Concourse${RESET}\n%s\n" "${name}" "${diff_output}")
      else
        if command -v fly > /dev/null; then
          fly validate-pipeline -c concourse.actual.yml &> /dev/null
        fi
      fi
    fi

    cd - > /dev/null
  fi

  if [[ ${failed} -eq 0 ]]; then
    printf "${GREEN}  pass${RESET}  %s\n" "${name}"
    echo "pass" > "${RESULTS_DIR}/${name}"
  else
    printf "${RED}  FAIL${RESET}  %s\n" "${name}"
    echo "fail" > "${RESULTS_DIR}/${name}"
    echo "${output}"
  fi
}

export HALFPIPE HALFPIPE_BRANCH DIFF_COLOR RESULTS_DIR
export RED GREEN BOLD RESET
export -f runTest

# Collect test directories, optionally filtered by argument
if [[ $# -gt 0 ]]; then
  dirs=()
  for pattern in "$@"; do
    for d in "${E2E_DIR}"/${pattern}/; do
      [[ -d "$d" ]] && dirs+=("$d")
    done
  done
  if [[ ${#dirs[@]} -eq 0 ]]; then
    echo "No test directories matched: $*"
    exit 1
  fi
else
  dirs=("${E2E_DIR}"/*/)
fi

# Run tests
if command -v parallel > /dev/null; then
  printf '%s\n' "${dirs[@]}" | parallel -k -j16 runTest
else
  for dir in "${dirs[@]}"; do
    runTest "${dir}"
  done
fi

# Summarise results
passed=0
failed=0
failed_names=()

for result_file in "${RESULTS_DIR}"/*; do
  [[ -f "${result_file}" ]] || continue
  name=$(basename "${result_file}")
  if [[ $(cat "${result_file}") == "pass" ]]; then
    ((passed++))
  else
    ((failed++))
    failed_names+=("${name}")
  fi
done

rm -rf "${RESULTS_DIR}"

echo ""
if [[ ${failed} -eq 0 ]]; then
  printf "${GREEN}${BOLD}All %d tests passed${RESET}\n" "${passed}"
  exit 0
else
  printf "${RED}${BOLD}%d of %d tests failed${RESET}\n" "${failed}" $((passed + failed))
  exit 1
fi
