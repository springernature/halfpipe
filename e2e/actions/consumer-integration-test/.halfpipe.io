team: halfpipe-team
pipeline: pipeline-name
platform: actions

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
  vars:
    K: value
    K1: value1

