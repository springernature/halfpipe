team: halfpipe-team
pipeline: pipeline-name

triggers:
  - type: git
    watched_paths:
      - e2e/actions/deploy-cf

tasks:
  - type: run
    docker:
      image: ubuntu
    name: make binary
    script: \echo foo > foo.html
    save_artifacts:
      - foo.html

  - type: deploy-cf
    name: deploy to cf
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest.yml
    deploy_artifact: foo.html
    vars:
      ENV1: 1234
      ENV2: ((secret.value))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))

  - type: deploy-cf
    name: deploy to cf with pre-promote
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest.yml
    deploy_artifact: foo.html
    vars:
      ENV2: ((secret.value))
      ENV4: ((another.secret))
    pre_promote:
      - type: run
        docker:
          image: alpine
        script: smoke-test.sh
        vars:
          ENV5: ((some.secret))
      - type: docker-compose
      - type: consumer-integration-test
        name: CDCs
        consumer: repo/app
        consumer_host: consumer.host
        script: ci/run-external-and-cdcs-dev
  - type: deploy-cf
    name: deploy to cf with docker image
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest-docker.yml
