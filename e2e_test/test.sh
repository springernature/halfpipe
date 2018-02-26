#!/usr/bin/env bash
set -e
set -o pipefail

HALFPIPE_PATH=${1:-"../halfpipe"}

${HALFPIPE_PATH} | tee pipeline.yml

fly validate-pipeline -c pipeline.yml
