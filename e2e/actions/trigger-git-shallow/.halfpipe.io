team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  shallow: false

tasks:
- type: run
  docker:
    image: alpine
  script: \date
- type: run
  docker:
    image: alpine
  script: \date
