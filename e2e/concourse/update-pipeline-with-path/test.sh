#!/usr/bin/env bash

../../../halfpipe -q -i .myCustomHalfpipePath.yml > pipelineActual.yml

diff --ignore-blank-lines pipelineActual.yml pipelineExpected.yml
