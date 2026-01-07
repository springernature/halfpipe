team: halfpipe-team
pipeline: halfpipe-e2e-run
notifications:
  failure:
  - slack: "#test"

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
  vars:
    MULTIPLE: ((levels/secret/deep.secret))
