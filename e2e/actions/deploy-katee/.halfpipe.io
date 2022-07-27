team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
  - type: git
    watched_paths:
      - e2e/actions/deploy-katee

tasks:
  - type: docker-push
    name: Push default
    image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
    tag: version
    image_scan_severity: low

  - type: deploy-katee
    name: deploy to katee
    image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
    tag: version
    applicationName: BLAHBLAH
    slackChannel: "#ee-re"
    velaAppFile: vela.yaml
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))