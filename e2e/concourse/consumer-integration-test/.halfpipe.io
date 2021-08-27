team: halfpipe-team
pipeline: halfpipe-e2e-consumer-integration-test

triggers:
- type: git
  watched_paths:
  - e2e/concourse/consumer-integration-test

tasks:
- type: consumer-integration-test
  name: another-c-name
  consumer: c-consumer
  consumer_host: c-host
  provider_host: p-host
  provider_name: p-name
  script: c-script
  docker_compose_service: potato
  git_clone_options: "--depth 100"
  use_covenant: true
  vars:
    K: value
    K1: value1
