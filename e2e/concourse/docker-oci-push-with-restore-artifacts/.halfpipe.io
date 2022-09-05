team: halfpipe-team
pipeline: docker-push-with-update-pipeline

feature_toggles:
  - update-pipeline

triggers:
  - type: git
    watched_paths:
      - e2e/concourse/docker-oci-push-with-restore-artifacts

tasks:
  - type: run
    docker:
      image: alpine
    script: ./build.sh
    save_artifacts:
      - file1

  - type: docker-push
    name: push to docker registry
    username: rob
    password: verysecret
    image: springerplatformengineering/image1
    restore_artifacts: true
    vars:
      A: a
      B: b

  - type: docker-push
    name: push to docker registry with git ref
    username: rob
    password: verysecret
    image: springerplatformengineering/image2
    restore_artifacts: true
    vars:
      A: a
      B: b
    tag: "file:some/folder/dockerTagFile"
