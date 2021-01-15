team: halfpipe-team
pipeline: pipeline-name

triggers:
- type: git
  watched_paths:
  - e2e/actions/deploy-cf

tasks:
- type: deploy-cf
  name: deploy to dev
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  cli_version: cf7

