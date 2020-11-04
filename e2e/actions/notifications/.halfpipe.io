team: halfpipe-team
pipeline: pipeline-name

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
    on_failure_message: failure message
