team: engineering-enablement
pipeline: halfpipe-e2e-timer-trigger

triggers:
- type: timer
  cron: "* * * * *"

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

