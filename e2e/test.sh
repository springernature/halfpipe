#!/usr/bin/env bash
runTest() {
  dir=${1%*/}
  echo "* Running ${dir}"
  mkdir -p /tmp/halfpipe-e2e/$dir
  tmpYml="/tmp/halfpipe-e2e/$dir.yml"

  cd ${dir}
  if [[ -f test.sh ]]; then
    ./test.sh
  elif [[ -f workflowExpected.yml ]]; then
    # actions
    ../../../halfpipe -q -o workflowActual.yml
    sed '6s/\"\"/main/' workflowActual.yml > $tmpYml
    diff --ignore-blank-lines $tmpYml workflowExpected.yml
  elif [[ -f pipelineExpected.yml ]]; then
    # concourse
    ../../../halfpipe -q -o pipelineActual.yml
    sed 's/    branch: ""/    branch: main/g' pipelineActual.yml | sed -E 's/(key:.+)\-$/\1/g' > $tmpYml
    diff --ignore-blank-lines $tmpYml pipelineExpected.yml
    if command -v fly > /dev/null; then
      fly validate-pipeline -c pipelineActual.yml &> /dev/null
    fi
  fi
}


rm -rf /tmp/halfpipe-e2e
export -f runTest

if command -v parallel > /dev/null; then
  ls -d */*/ | parallel -j16 runTest
else
  # xargs doesn't return exit code reliably with parallel
  ls -d */*/ | xargs -ID -P1 bash -c "runTest D"
fi
