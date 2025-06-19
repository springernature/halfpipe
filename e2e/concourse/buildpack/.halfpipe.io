team: halfpipe-team
pipeline: halfpipe-e2e-buildpack

triggers:
  - type: git
    watched_paths:
      - e2e/concourse/buildpack

tasks:
  - type: buildpack
    name: create-docker-image
    path: build/libs
    buildpacks: gcr.io/paketo-buildpacks/java:18.5.0,gcr.io/paketo-buildpacks/node:18.5.0
    image: eu.gcr.io/halfpipe-io/halfpipe-e2e-pact
