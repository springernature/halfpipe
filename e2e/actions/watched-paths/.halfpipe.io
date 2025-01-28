team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  watched_paths:
  - go.*
  - build.sh

tasks:
- type: run
  name: build
  docker:
    image: foo
  script: \build
