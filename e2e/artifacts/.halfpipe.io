team: main
pipeline: halfpipe-e2e-artifacts

triggers:
- type: git
  branch: 6.5.1
  watched_paths:
  - e2e/artifacts

tasks:
- type: run
  name: create-artifact
  script: ./a
  docker:
    image: alpine
  save_artifacts:
  - someFile
  - ../parentDir/someFile2
  save_artifacts_on_failure:
  - .halfpipe.io.yml
  - ../../.halfpipe.io.yml

- type: run
  name: read-artifact
  script: ./a
  docker:
    image: alpine
  restore_artifacts: true
