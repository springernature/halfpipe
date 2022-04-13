team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  shallow: false
  git_crypt_key: ((foo.bar))

tasks:
- type: run
  docker:
    image: alpine
  script: \date
  vars:
    FOO: bar
- type: run
  docker:
    image: alpine
  script: \date
