#!/usr/bin/env bash
set -e

#(cd concourse; ./test.sh "${1-}")
(cd actions; ./test.sh "${1-}")

