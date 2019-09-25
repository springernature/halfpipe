team: engineering-enablement
pipeline: halfpipe-e2e-parallel-without-update-job

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/parallel-without-update-job

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
  name: not parallel
  script: ./a
  docker:
    image: alpine:test
  parallel: false

- type: parallel
  tasks:
  - type: run
    name: test parallel 5
    script: ./a
    docker:
      image: alpine:test

  - type: run
    name: test parallel 6
    script: ./a
    docker:
      image: alpine:test

  - type: run
    name: test parallel 7
    script: ./a
    docker:
      image: alpine:test

- type: run
  name: one group
  script: ./a
  docker:
    image: alpine:test

- type: run
  name: after parallel
  script: ./a
  docker:
    image: alpine:test
