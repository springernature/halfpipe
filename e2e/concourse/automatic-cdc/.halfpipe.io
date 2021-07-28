team: halfpipe-team
pipeline: halfpipe-e2e-automatic-cdc

triggers:
- type: git
  watched_paths:
  - e2e/concourse/automatic-cdc

tasks:
- type: automatic-cdc
  name: Automatic CDCs