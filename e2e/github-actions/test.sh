#!/usr/bin/env bash
set -e

../../halfpipe 1> pipeline.yml
diff pipeline.yml expected-pipeline.yml
