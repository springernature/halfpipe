team: halfpipe-team
pipeline: pipeline-name

triggers:
- type: git
  watched_paths:
  - e2e/actions/docker-push
  - e2e/actions
  ignored_paths:
  - README.md
  - '**.js'
