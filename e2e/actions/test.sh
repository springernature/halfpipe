#!/usr/bin/env bash
set -e

for dir in */
do
    dir=${dir%*/}
    echo "  * Running actions/${dir}"
    (
        cd ${dir}
        if [[ -f test.sh ]]; then
            ./test.sh
        else
            ../../../halfpipe -q actions 1> workflowActual.yml
            sed '6s/\"\"/master/' workflowActual.yml > /tmp/branchFixed.yml
            diff --ignore-blank-lines /tmp/branchFixed.yml workflowExpected.yml
        fi
    )
done
