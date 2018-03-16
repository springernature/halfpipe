#!/usr/bin/env bash

# Update pipeline.yml and push to Concourse
# $1 pipeline-name (optional - defaults to current directory name)
hp() {
  echo Updating pipeline.yml
  halfpipe > pipeline.yml
  echo Uploading to Concourse
  fly-set $1
}

# Login to Concourse
# $1 team-name (optional - defaults to engineering-enablement)
fly-login() {
  team=${1:-"engineering-enablement"}
  fly -t hp login -n $team
}

# Get a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
fly-get() {
  dir=$(basename "$PWD")
  pipeline=${1:-$dir}
  fly -t hp get-pipeline -p $pipeline
}

# Set a pipeline
# $1 pipeline-name (optional - defaults to current directory name)
# $2 config-yaml   (optional - defaults to pipeline.yml)
fly-set() {
  dir=$(basename "$PWD")
  pipeline=${1:-$dir}
  config=${2:-pipeline.yml}
  setp="fly -t hp set-pipeline -p $pipeline -c $config"
  echo ">> $setp"
  $setp
}
