#!/usr/bin/env bash

cat halfpipe.io | ../../halfpipe > pipeline.yml
diff pipeline.yml expected-pipeline.yml

