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
  - type: docker-compose
    name: C
