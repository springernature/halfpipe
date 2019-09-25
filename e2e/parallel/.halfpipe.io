team: engineering-enablement
pipeline: halfpipe-e2e-parallel

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/parallel

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: parallel 1 1
  script: ./a
  docker:
    image: alpine
  parallel: 1

- type: run
  name: parallel 1 2
  script: ./a
  docker:
    image: alpine
  parallel: 1

- type: run
  name: parallel 2 1
  script: ./a
  docker:
    image: alpine
  parallel: blah

- type: run
  name: parallel 2 2
  script: ./a
  docker:
    image: alpine
  parallel: blah

- type: run
  name: not parallel
  script: ./a
  docker:
    image: alpine
  parallel: false

- type: parallel
  tasks:
  - type: run
    name: parallel 3 1
    script: ./a
    docker:
      image: alpine

  - type: run
    name: parallel 3 2
    script: ./a
    docker:
      image: alpine

  - type: run
    name: parallel 3 3
    script: ./a
    docker:
      image: alpine

- type: parallel
  tasks:
  - type: sequence
    tasks:
    - type: run
      name: parallel 4 sequence 1 1
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 4 sequence 1 2
      script: ./a
      docker:
        image: alpine
  - type: sequence
    tasks:
    - type: run
      name: parallel 4 sequence 2 1
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 4 sequence 2 2
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 4 sequence 2 3
      script: ./a
      docker:
        image: alpine
  - type: run
    name: parallel 4 1
    script: ./a
    docker:
      image: alpine

- type: run
  name: after parallel
  script: ./a
  docker:
    image: alpine
