team: test
pipeline: test

triggers:
- type: git
  manual_trigger: true
tasks:
- type: run
  name: test
  script: ./a
  docker:
    image: alpine:test
