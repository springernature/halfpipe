team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: run
  name: my run task
  docker:
    image: foo
  script: \foo
  notifications:
    success:
    - slack: "#success1"
      message: success message
    - slack: "#success2"
      message: success message
    - teams: http://webhook-success
      message: success message teams
    failure:
    - slack: "#failure1"
    - teams: ((secret.webhook1))
    - teams: http://webhook2
