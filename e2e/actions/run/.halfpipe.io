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
