#!/usr/bin/env bash

../../../halfpipe -i .myCustomHalfpipePath.yml > pipelineActual.yml

# hacky fixes for running tests on a branch
sed 's/    branch: ""/    branch: main/g' pipelineActual.yml > /tmp/x && mv /tmp/x pipelineActual.yml
sed -E 's/(key:.+)\-$/\1/g' pipelineActual.yml > /tmp/x && mv /tmp/x pipelineActual.yml

diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
