team: test
pipeline: test
repo:
  shallow: true
  watched_paths:
  - e2e/parallel
tasks:
- type: run
  name: test
  script: ./a
  docker:
    image: alpine:test

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
  name: after parallel
  script: ./a
  docker:
    image: alpine:test
