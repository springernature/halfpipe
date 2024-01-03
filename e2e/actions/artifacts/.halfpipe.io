team: halfpipe-team
pipeline: artifacts
platform: actions

triggers:
- type: git
  watched_paths:
  - e2e/actions/artifacts

tasks:
- type: run
  name: build
  docker:
    image: debian:buster-slim
  script: \mkdir target; echo foo > foo.txt; echo bar > target/bar.txt
  save_artifacts:
  - foo.txt
  - target/bar.txt
  - ../test.sh
  save_artifacts_on_failure:
  - foo.txt

- type: run
  name: test
  docker:
    image: debian:buster-slim
  script: \ls -l ..; cat foo.txt target/bar.txt ../test.sh
  restore_artifacts: true
