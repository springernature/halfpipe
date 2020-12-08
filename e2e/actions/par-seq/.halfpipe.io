team: halfpipe-team
pipeline: par-seq

tasks:
- type: run
  name: task 1
  docker:
    image: alpine:latest
  script: \date
- type: run
  name: task 2
  docker:
    image: alpine:latest
  script: \date
- type: run
  name: task 3
  docker:
    image: alpine:latest
  script: \date
- type: parallel
  tasks:
  - type: run
    name: task 4.1
    docker:
      image: alpine:latest
    script: \date
  - type: run
    name: task 4.2
    docker:
      image: alpine:latest
    script: \date
  - type: sequence
    tasks:
    - type: run
      name: task 4.3.1
      docker:
        image: alpine:latest
      script: \date
    - type: run
      name: task 4.3.2
      docker:
        image: alpine:latest
      script: \date
- type: run
  name: task 5
  docker:
    image: alpine:latest
  script: \date
