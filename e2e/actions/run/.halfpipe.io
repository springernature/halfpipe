team: halfpipe-team
pipeline: halfpipe-e2e-run

triggers:
- type: git
  watched_paths:
  - e2e/actions/run

tasks:
- type: run
  name: This is a test
  script: ./a
  docker:
    image: alpine:test
