team: halfpipe-team
pipeline: halfpipe-e2e-parallel

triggers:
- type: git
  shallow: true
  watched_paths:
  - e2e/concourse/parallel

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: first job
  script: ./a
  docker:
    image: alpine

- type: parallel
  tasks:
  - type: run
    name: parallel 1 1
    script: ./a
    docker:
      image: alpine

  - type: run
    name: parallel 1 2
    script: ./a
    docker:
      image: alpine

  - type: run
    name: parallel 1 3
    script: ./a
    docker:
      image: alpine

- type: parallel
  tasks:
  - type: sequence
    tasks:
    - type: run
      name: parallel 2 sequence 1 1
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 2 sequence 1 2
      script: ./a
      docker:
        image: alpine
    - type: parallel
      tasks:
      - type: run
        name: parallel 2 sequence 1 2a
        script: ./a
        docker:
          image: alpine
      - type: run
        name: parallel 2 sequence 1 2b
        script: ./a
        docker:
          image: alpine
  - type: sequence
    tasks:
    - type: run
      name: parallel 2 sequence 2 1
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 2 sequence 2 2
      script: ./a
      docker:
        image: alpine
    - type: run
      name: parallel 2 sequence 2 3
      script: ./a
      docker:
        image: alpine
  - type: run
    name: parallel 2 1
    script: ./a
    docker:
      image: alpine

- type: run
  manual_trigger: true
  name: after parallel
  script: ./a
  docker:
    image: alpine
