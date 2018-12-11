team: test
pipeline: test
repo:
  watched_paths:
  - e2e/docker-compose

tasks:
- type: docker-compose
  name: test
