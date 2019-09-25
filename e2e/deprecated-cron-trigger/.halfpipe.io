team: engineering-enablement
pipeline: halfpipe-e2e-deprecated-cron-trigger

cron_trigger: "10 10 * * *"

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

