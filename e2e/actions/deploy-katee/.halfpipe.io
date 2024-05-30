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

  - type: deploy-katee
    name: deploy to katee
    tag: version
    deployment_check_timeout: 120
    notifications:
      on_failure:
        - "#ee-re"
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))
      VERY_SECRET: ((another.secret))

  - type: deploy-katee
    name: deploy to katee different team
    tag: version
    namespace: katee-different-namespace
    environment: katee-different-environment
    platform_version: v2
    notifications:
      on_failure:
        - "#ee-re"
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))
      VERY_SECRET: ((another.secret))
