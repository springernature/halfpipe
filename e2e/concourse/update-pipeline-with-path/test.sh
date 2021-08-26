#!/usr/bin/env bash

../../../halfpipe -i .myCustomHalfpipePath.yml > pipelineActual.yml

# hacky fixes for running tests on a branch
sed -i '' 's/    branch: ""/    branch: main/g' pipelineActual.yml
sed -i '' -E 's/(key:.+)\-$/\1/g' pipelineActual.yml

diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
