team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  ignored_paths:
  - README.md
  - '**.js'
