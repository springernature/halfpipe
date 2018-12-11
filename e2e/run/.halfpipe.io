team: test
pipeline: test
repo:
  watched_paths:
  - e2e/run
tasks:
- type: run
  name: test
  script: ./a
  docker:
    image: alpine:test
