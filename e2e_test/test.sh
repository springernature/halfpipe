#/usr/bin/env bash
set -euo pipefail

HALFPIPE_PATH=${1:-"../halfpipe"}

${HALFPIPE_PATH} | tee pipeline.yml

command -v fly >/dev/null && fly validate-pipeline -c pipeline.yml

diff -w pipeline.yml expected-pipeline.yml
