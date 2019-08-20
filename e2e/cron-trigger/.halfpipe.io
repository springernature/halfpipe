team: test
pipeline: test

triggers:
- type: cron
  trigger: "* * * * *"

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

