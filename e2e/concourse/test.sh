#!/usr/bin/env bash
set -e

for dir in */
do
    dir=${dir%*/}
    echo "  * Running concourse/${dir}"
    (
        cd ${dir}
        if [[ -f test.sh ]]; then
            ./test.sh
        else
            ../../../halfpipe -q 1> pipelineActual.yml

            # hacky fixes for running tests on a branch
            sed 's/    branch: ""/    branch: main/g' pipelineActual.yml > /tmp/x && mv /tmp/x pipelineActual.yml
            sed -E 's/(key:.+)\-$/\1/g' pipelineActual.yml > /tmp/x && mv /tmp/x pipelineActual.yml

            diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
            if command -v fly > /dev/null; then
                fly validate-pipeline -c pipelineActual.yml > /dev/null 2>&1
            fi
        fi
    )
done
