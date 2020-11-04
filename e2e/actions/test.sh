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
            ../../../halfpipe actions 1> workflowActual.yml
            diff --ignore-blank-lines workflowActual.yml workflowExpected.yml
        fi
    )
done
