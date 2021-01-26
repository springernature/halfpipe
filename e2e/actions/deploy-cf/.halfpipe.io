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
