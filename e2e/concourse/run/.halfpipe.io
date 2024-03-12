team: halfpipe-team
pipeline: halfpipe-e2e-run

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/concourse/run
tasks:
- type: run
  name: test
  script: ./a
  privileged: false
  docker:
    image: alpine:test
  build_history: 10
  vars:
    MULTIPLE: ((levels/secret/deep.secret))
