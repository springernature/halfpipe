team: halfpipe-team
pipeline: pipeline-name

tasks:
- type: run
  name: create zip for ml task
  script: \package.sh
  docker:
    image: alpine
  save_artifacts:
  - xquery.zip

- type: deploy-ml-zip
  deploy_zip: target/xquery.zip
  use_build_version: true
  targets:
  - ml.dev.com

- type: deploy-ml-modules
  name: Deploy ml-modules artifact
  ml_modules_version: "2.1425"
  app_name: my-app
  app_version: v1
  use_build_version: false
  targets:
  - ml.dev.com
  - ml.qa.com
  username: foo
