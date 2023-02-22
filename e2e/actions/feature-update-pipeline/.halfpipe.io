team: team
pipeline: update-pipeline
platform: actions

feature_toggles:
- update-pipeline

tasks:
- type: docker-compose
  name: A
- type: parallel
  tasks:
  - type: docker-compose
    name: B
    vars:
      S: ((very.secret))
  - type: docker-compose
    name: C
    vars:
      S1: ((very.secret1))
