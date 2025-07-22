team: halfpipe-team
pipeline: halfpipe-e2e-buildpack
platform: actions

triggers:
  - type: git
    watched_paths:
      - e2e/concourse/buildpack

tasks:
  - type: buildpack
    name: pack-n-push
    path: build/libs
    buildpacks: gcr.io/paketo-buildpacks/java:18.5.0,gcr.io/paketo-buildpacks/node:18.5.0
    image: eu.gcr.io/halfpipe-io/engineering-enablement/halfpipe-e2e-buildpack
    vars:
      BP_FOO: foo
      BP_BAR: bar
