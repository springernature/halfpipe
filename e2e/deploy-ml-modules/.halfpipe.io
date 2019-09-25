team: engineering-enablement
pipeline: halfpipe-e2e-deploy-ml-modules

triggers:
- type: git
  watched_paths:
  - e2e/deploy-ml-modules

tasks:
- type: deploy-ml-modules
  name: Deploy ml-modules artifact
  ml_modules_version: "2.1425"
  app_name: my-app
  app_version: v1
  use_build_version: false
  targets:
  - ml.dev.springer-sbm.com
  - ml.qa1.springer-sbm.com
