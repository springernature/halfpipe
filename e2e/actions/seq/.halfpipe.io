team: halfpipe-team
pipeline: pipeline-name

tasks:
- type: run
  name: task 1
  docker:
    image: img:tag
  script: \script
- type: run
  name: task 2
  docker:
    image: img:tag
  script: \script
- type: run
  name: task 3
  docker:
    image: img:tag
  script: \script
