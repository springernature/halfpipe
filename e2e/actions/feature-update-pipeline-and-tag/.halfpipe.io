team: team
pipeline: update-pipeline-and-tag

feature_toggles:
- update-pipeline-and-tag

tasks:
- type: docker-compose
  name: A
- type: parallel
  tasks:
  - type: docker-compose
    name: B
  - type: docker-compose
    name: C
