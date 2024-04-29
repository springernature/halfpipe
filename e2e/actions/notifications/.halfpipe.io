team: halfpipe-team
pipeline: pipeline-name
platform: actions

slack_channel: "#channel"
slack_failure_message: failure msg # This will not be set in the notification on the task, since notifications is already defined on the task. I.e no complicated merging!
teams_webhook: http://blah # has no effect because the task has old style notification block.

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
