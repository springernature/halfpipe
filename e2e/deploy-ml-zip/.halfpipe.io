team: test
pipeline: test

triggers:
- type: git
  watched_paths:
  - e2e/deploy-ml-zip

tasks:
- type: run
  name: create zip for ml task
  script: a
  docker:
    image: alpine
  save_artifacts:
  - ml.zip

- type: deploy-ml-zip
  deploy_zip: target/xquery.zip
  use_build_version: true
  targets:
  - ml.dev.springer-sbm.com

