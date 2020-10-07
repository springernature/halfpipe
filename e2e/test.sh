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
            ../../halfpipe 1> pipeline.yml
            diff --ignore-blank-lines pipeline.yml expected-pipeline.yml
            if command -v fly > /dev/null; then
                fly validate-pipeline -c pipeline.yml > /dev/null
            fi
        fi
    )
done
