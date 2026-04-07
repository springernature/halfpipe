#!/usr/bin/env bash
runTest() {
  dir=${1%*/}
  echo "* Running ${dir}"

  cd ${dir}
  if [[ -f test.sh ]]; then
    ./test.sh
  elif [[ -f workflowExpected.yml ]]; then
    # actions
    ../../../halfpipe -q -o workflowActual.yml
    diff --ignore-blank-lines workflowActual.yml workflowExpected.yml
  elif [[ -f pipelineExpected.yml ]]; then
    # concourse
    ../../../halfpipe -q -o pipelineActual.yml
    diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
    if command -v fly > /dev/null; then
      fly validate-pipeline -c pipelineActual.yml &> /dev/null
    fi
  fi
}


export HALFPIPE_BRANCH=main
export -f runTest

if command -v parallel > /dev/null; then
  ls -d */*/ | parallel -j16 runTest
else
  # xargs doesn't return exit code reliably with parallel
  ls -d */*/ | xargs -ID -P1 bash -c "runTest D"
fi
