team: halfpipe-team
pipeline: pipeline-name

triggers:
  - type: git
    watched_paths:
      - e2e/actions/deploy-katee

feature_toggles:
- update-pipeline

tasks:
  - type: docker-push
    name: Push default
    image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
    tag: version

  - type: deploy-katee
    name: deploy to katee
    image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
    tag: version
    application_name: BLAHBLAH
    vela_manifest: vela.yaml
    notifications:
      on_failure:
        - "#ee-re"
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))
      VERY_SECRET: blah

  - type: deploy-katee
    name: deploy to katee different team
    image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
    tag: version
    team: different-team
    application_name: BLAHBLAH
    vela_manifest: vela.yaml
    notifications:
      on_failure:
        - "#ee-re"
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))
      VERY_SECRET: blah
