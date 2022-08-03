#!/usr/bin/env bash
runTest() {
  dir=${1%*/}
  yml="/tmp/halfpipe-e2e/$dir.yml"
  log="/tmp/halfpipe-e2e/$dir.log"
  (
    set -e
    echo "* Running ${dir}"
    cd ${dir}
    if [[ -f test.sh ]]; then
      ./test.sh
    elif [[ -f workflowExpected.yml ]]; then
      # actions
      ../../../halfpipe -q -o workflowActual.yml
      sed '6s/\"\"/main/' workflowActual.yml > $yml
      diff --ignore-blank-lines $yml workflowExpected.yml
    else
      # concourse
      ../../../halfpipe -q -o pipelineActual.yml
      sed 's/    branch: ""/    branch: main/g' pipelineActual.yml | sed -E 's/(key:.+)\-$/\1/g' > $yml
      diff --ignore-blank-lines $yml pipelineExpected.yml
      if command -v fly > /dev/null; then
        fly validate-pipeline -c pipelineActual.yml &> /dev/null
      fi
    fi
  ) &>> $log
}

export -f runTest

rm -rf /tmp/halfpipe-e2e
mkdir -p /tmp/halfpipe-e2e/{actions,concourse}
ls -d */*/ | xargs -ID -P16 bash -c "runTest D"
RET_CODE=$?
cat /tmp/halfpipe-e2e/*/*.log
exit $RET_CODE
