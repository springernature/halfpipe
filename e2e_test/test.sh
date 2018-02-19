set -e

../halfpipe | tee pipeline.yml

fly validate-pipeline -c pipeline.yml
