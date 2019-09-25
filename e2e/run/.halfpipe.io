team: engineering-enablement
pipeline: halfpipe-e2e-run

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/run
tasks:
- type: run
  name: test
  script: ./a
  privileged: false
  docker:
    image: alpine:test
