team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  watched_paths:
  - e2e/actions/watched-paths/main.*
  - e2e/actions/watched-paths/build-project.sh

tasks:
- type: run
  name: build
  docker:
    image: foo
  script: \build
