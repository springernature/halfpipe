#!/usr/bin/env bash
set -e

for dir in */
do
    dir=${dir%*/}
    echo "  * Running ${dir}"
    (
        cd ${dir}
        if [[ -f test.sh ]]; then
            ./test.sh
        else
            ../../halfpipe 1> pipeline.yml 2> /dev/null
            diff -w pipeline.yml expected-pipeline.yml
            if [ "${1-}" != "ci" ]; then
                command -v fly >/dev/null && fly validate-pipeline -c pipeline.yml > /dev/null
            fi
        fi
    )
done