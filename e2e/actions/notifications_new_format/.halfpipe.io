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
    slack:
      on_success:
      - '#success1'
      - '#success2'
      on_success_message: success message
      on_failure:
      - '#failure1'
    teams:
      on_failure:
      - 'http://webhook1'
      - 'http://webhook2'
      on_success:
      - 'http://webhook-success'
      on_success_message: success message teams
