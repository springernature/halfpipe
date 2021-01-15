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
    api: dev-api
    space: dev
    manifest: manifest.yml
    username: michiel
    password: very-secret
    deploy_artifact: foo.html
    test_domain: some.random.domain.com
