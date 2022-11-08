team: halfpipe-team
pipeline: halfpipe-e2e-run

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/concourse/run

feature_toggles:
- github-statuses

tasks:
- type: run
  name: test
  script: ./a
  privileged: false
  docker:
    image: alpine:test

- type: run
  name: test again
  script: ./a
  privileged: false
  docker:
    image: alpine:test
