team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
  - type: git
    watched_paths:
      - e2e/actions/deploy-cf

tasks:
  - type: run
    docker:
      image: ubuntu
    name: make binary
    script: \echo foo > foo.html
    save_artifacts:
      - foo.html

  - type: deploy-cf
    name: deploy to cf
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest.yml
    deploy_artifact: foo.html
    github_environment:
      name: my-env
      url: https://my-url
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))

  - type: deploy-cf
    name: deploy to cf with cf8
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest.yml
    deploy_artifact: foo.html
    cli_version: cf8
    vars:
      ENV1: 1234
      ENV2: ((secret.something))
      ENV3: '{"a": "b", "c": "d"}'
      ENV4: ((another.secret))

  - type: deploy-cf
    name: deploy to cf with pre-promote
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest.yml
    deploy_artifact: foo.html
    vars:
      ENV2: ((secret.something))
      ENV4: ((another.secret))
    pre_promote:
      - type: run
        docker:
          image: alpine
        script: smoke-test.sh
        vars:
          ENV5: ((some.secret))
      - type: docker-compose
      - type: consumer-integration-test
        name: CDCs
        consumer: repo/app
        consumer_host: consumer.host
        script: ci/run-external-and-cdcs-dev
  - type: deploy-cf
    name: deploy to cf with docker image
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    manifest: manifest-docker.yml
    docker_tag: version

  - type: deploy-cf
    name: deploy with sso
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    deploy_artifact: foo.html
    sso_route: my-route.public.springernature.app
    rolling: true

  - type: deploy-cf
    name: deploy without artifact
    api: ((cloudfoundry.api-snpaas))
    org: ((cloudfoundry.org-snpaas))
    space: dev
    sso_route: my-route.public.springernature.app
    rolling: true
