#!/usr/bin/env bash

# put something like this in your ~/.bash_profile
# [ -f ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh ] && source ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh

# concourse.halfpipe.io default target is "hp"
target=${FLY_TARGET:-hp}


# Update pipeline.yml and push to Concourse
# $1 pipeline-name (optional - defaults to current directory name)
hp() {
  echo Updating pipeline.yml
  halfpipe > pipeline.yml &&
    echo Uploading to Concourse &&
      hp-set $1
}

# Login to Concourse
# $1 team-name (optional - defaults to engineering-enablement)
hp-login() {
  team=${1:-"$(grep 'team:' .halfpipe.io | cut -d: -f2)"}
  login="fly -t ${target} login -n ${team}"
  echo ">> ${login}"
  ${login}
}

# Get a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
hp-get() {
  pipeline=${1:-"$(grep 'pipeline:' .halfpipe.io | cut -d: -f2)"}
  getPipeline="fly -t $target get-pipeline -p $pipeline"
  echo ">> ${getPipeline}"
  ${getPipeline}
}

# Set a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
# $2 config-yaml   (optional - defaults to pipeline.yml)
hp-set() {
  pipeline=${1:-"$(grep 'pipeline:' .halfpipe.io | cut -d: -f2)"}
  config=${2:-pipeline.yml}
  setPipeline="fly -t $target set-pipeline -p $pipeline -c $config"
  echo ">> ${setPipeline}"
  ${setPipeline}
}
