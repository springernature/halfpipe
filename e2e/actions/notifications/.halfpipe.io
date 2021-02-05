team: halfpipe-team
pipeline: pipeline-name
slack_failure_message: failure msg

tasks:
- type: run
  name: my run task
  docker:
    image: foo
  script: \foo
  notifications:
    on_success:
    - '#success1'
    - '#success2'
    on_success_message: success message
    on_failure:
    - '#failure1'
