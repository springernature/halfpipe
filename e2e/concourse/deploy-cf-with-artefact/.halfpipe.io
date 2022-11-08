team: halfpipe-team
pipeline: halfpipe-e2e-deploy-cf

triggers:
- type: git
  watched_paths:
  - e2e/concourse/deploy-cf

tasks:
- type: run
  docker:
    image: ubuntu
  name: make binary
  script: \make
  save_artifacts:
  - build/linux/binary

- type: deploy-cf
  name: deploy to cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  timeout: 5m
  deploy_artifact: build/linux/binary

