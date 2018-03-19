#!/usr/bin/env bash

# put something like this in your ~/.bash_profile
# [ -f ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh ] && source ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh

# assumes concourse.halfpipe.io is setup as target "hp". don't argue.
readonly target=hp

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
  team=${1:-"engineering-enablement"}
  fly -t $target login -n $team
}

# Get a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
hp-get() {
  dir=$(basename "$PWD")
  pipeline=${1:-$dir}
  fly -t $target get-pipeline -p $pipeline
}

# Set a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
# $2 config-yaml   (optional - defaults to pipeline.yml)
hp-set() {
  dir=$(basename "$PWD")
  pipeline=${1:-$dir}
  config=${2:-pipeline.yml}
  setp="fly -t $target set-pipeline -p $pipeline -c $config"
  echo ">> $setp"
  $setp
}
