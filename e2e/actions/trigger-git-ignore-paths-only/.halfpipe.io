team: halfpipe-team
pipeline: pipeline-name
output: actions

triggers:
- type: git
  ignored_paths:
  - README.md
  - '**.js'
