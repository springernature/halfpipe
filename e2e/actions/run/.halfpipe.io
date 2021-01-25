team: halfpipe-team
pipeline: pipeline-name

tasks:
- type: run
  name: my run task
  docker:
    image: eu.gcr.io/halfpipe-io/golang:1.15
  script: \foo
  vars:
    FOO: foo
    BAR: bar
    SECRET1: ((something.cool))
    SECRET2: ((something.cooler))
  timeout: 1h2m
- type: run
  name: my run script
  docker:
    image: eu.gcr.io/halfpipe-io/golang:1.15
  script: \bash -c "echo hello"