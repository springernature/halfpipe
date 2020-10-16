team: halfpipe-team
pipeline: halfpipe-e2e-manual-git-trigger

triggers:
- type: git
  branch: 6.5.1
  manual_trigger: true

tasks:
- type: run
  name: test
  script: ./a
  docker:
    image: alpine:test
