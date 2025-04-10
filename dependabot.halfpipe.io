# the build script generates a github worfklow from this so
# that dependabot can check all the 3rd party actions we use

team: halfpipe-team
pipeline: for-dependabot-to-check
platform: actions
slack_channel: '#halfpipe-dev'

triggers:
- type: git
  manual_trigger: true

tasks:
- type: run
  script: \exit 1
  docker:
    image: eu.gcr.io/halfpipe-io/halfpipe-team/blah:nonexistent
  save_artifacts:
  - artifact.zip

- type: run
  script: \exit 1
  docker:
    image: eu.gcr.io/halfpipe-io/halfpipe-team/blah:nonexistent
  restore_artifacts: true

- type: deploy-cf
  api: ((cloudfoundry.api-snpaas))
  space: cf-space
  manifest: e2e/actions/deploy-cf/manifest.yml

- type: docker-push
  image: eu.gcr.io/halfpipe-io/halfpipe-team/blah
  dockerfile_path: e2e/actions/docker-push/Dockerfile

- type: docker-compose
  compose_file: e2e/actions/docker-compose/docker-compose.yml
