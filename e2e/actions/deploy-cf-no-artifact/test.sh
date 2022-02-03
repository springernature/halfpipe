#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# all this stupid stuff because the test requires the
# halfpipe manifest be in the root of the git project

rm -rf /tmp/halfpipe-test
cp -r ${SCRIPT_DIR} /tmp/halfpipe-test
cd /tmp/halfpipe-test
git init > /dev/null
git add test.sh > /dev/null
git commit -am"init" > /dev/null
git remote add origin https://github.com/foo/bar.git > /dev/null

halfpipe -q -o workflowActual.yml
sed '6s/\"\"/main/' workflowActual.yml > /tmp/branchFixed.yml
diff --ignore-blank-lines /tmp/branchFixed.yml workflowExpected.yml
