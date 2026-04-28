# yaml-language-server: $schema=../../.generated/schema.json
team: halfpipe-team
pipeline: pipeline-name

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: run
  script: ./test.sh
  docker:
    image: alpine:test
  vars:
    ENV1: 1234
    ENV2: ((secret.something))
    ENV3: '{"a": "b", "c": "d"}'
    ENV4: ((another.secret))
    VERY_SECRET: blah

- type: docker-compose
  name: docker-compose-simple

- type: docker-compose
  name: docker-compose-complex
  command: \echo hello
  compose_file: custom-docker-compose.yml docker-compose.yml
  service: customservice
  vars:
    ENV1: 1234
    ENV2: ((secret.something))
    ENV3: '{"a": "b", "c": "d"}'
    ENV4: ((another.secret))
    VERY_SECRET: blah

- type: buildpack
  name: buildpack
  image: eu.gcr.io/halfpipe-io/halfpipe-team/myapp
  buildpacks:
  - paketo-buildpacks/java
  - paketo-buildpacks/procfile
  vars:
    ENV1: 1234
    ENV2: ((secret.something))
    ENV3: '{"a": "b", "c": "d"}'
