set -e
set -o pipefail

HALFPIPE_PATH=${1:-"../halfpipe_linux"}

${HALFPIPE_PATH} | tee pipeline.yml

fly validate-pipeline -c pipeline.yml

diff -w pipeline.yml expected-pipeline.yml