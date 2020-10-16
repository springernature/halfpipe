team: halfpipe-team
pipeline: halfpipe-e2e-run

triggers:
- type: git
  shallow: true
  branch: 6.5.1
  watched_paths:
  - e2e/run
tasks:
- type: run
  name: test
  script: ./a
  privileged: false
  docker:
    image: alpine:test
  build_history: 10
