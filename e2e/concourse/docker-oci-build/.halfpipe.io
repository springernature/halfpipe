team: halfpipe-team
pipeline: halfpipe-e2e-docker-oci-build

feature_toggles:
- docker-oci-build

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-oci-build

tasks:
- type: docker-push
  name: push to docker registry
  username: uSeRnAmE
  password: verysecret
  image: springerplatformengineering/halfpipe_fly:thisIsMy_Tag

- type: run
  name: date to file
  docker:
    image: alpine
  script: \date > dateFile
  save_artifacts:
    - dateFile

- type: docker-push
  name: build-with-artifact-and-vars
  username: uSeRnAmE
  password: verysecret
  image: eu.gcr.io/halfpipe-io/engineering-enablement/simon-test-simple
  vars:
    PASSED_IN: SIMON
  restore_artifacts: true
  ignore_vulnerabilities: true

- type: docker-push
  name: different-build-and-file-folder
  username: uSeRnAmE
  password: verysecret
  image: eu.gcr.io/halfpipe-io/engineering-enablement/different-folders
  dockerfile_path: dockerFiles/randomDockerfile
  build_path: buildFolder
