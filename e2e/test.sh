#!/usr/bin/env bash
set -e

(echo Actions:; cd actions; ./test.sh "${1-}")
#(echo Concourse:; cd concourse; ./test.sh "${1-}")
