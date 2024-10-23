team: halfpipe-team
pipeline: halfpipe-e2e-timer-trigger

triggers:
- type: git
- type: timer
  cron: "0 20 * * 6"

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b
