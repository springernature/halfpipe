team: test
pipeline: test
repo:
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
  targets:
  - ml.dev.springer-sbm.com

