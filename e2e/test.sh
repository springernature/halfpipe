#!/usr/bin/env bash
set -e

(echo Concourse:; cd concourse; ./test.sh "${1-}")
(echo Actions:; cd actions; ./test.sh "${1-}")

