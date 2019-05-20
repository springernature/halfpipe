team: test
pipeline: test
repo:
  shallow: true
  watched_paths:
  - e2e/parallel

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: test parallel 1
  script: ./a
  docker:
    image: alpine:test
  parallel: true

- type: run
  name: test parallel 2
  script: ./a
  docker:
    image: alpine:test
  parallel: true

- type: run
  name: test parallel 3
  script: ./a
  docker:
    image: alpine:test
  parallel: blah

- type: run
  name: test parallel 4
  script: ./a
  docker:
    image: alpine:test
  parallel: blah

- type: run
  name: after parallel
  script: ./a
  docker:
    image: alpine:test
