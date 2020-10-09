#!/usr/bin/env bash
set -e

for system in */
do
  system=${system%*/}
  cmd=""
  if [[ "$system" = "actions" ]]; then cmd="actions"; fi
  echo "* Running ${system}"
  (
    cd $system
    for dir in */
    do
        dir=${dir%*/}
        echo "  * Running ${dir}"
        (
            cd ${dir}
            if [[ -f test.sh ]]; then
                ./test.sh
            else
                ../../../halfpipe $cmd 1> pipeline.yml
                diff --ignore-blank-lines pipeline.yml expected-pipeline.yml
                if command -v fly > /dev/null; then
                    fly validate-pipeline -c pipeline.yml > /dev/null
                fi
            fi
        )
    done
  )
done
