team: halfpipe-team
pipeline: pipeline-name
output: actions

triggers:
- type: git
  watched_paths:
  - e2e/actions/docker-push
  - e2e/actions
  ignored_paths:
  - README.md
  - '**.js'
